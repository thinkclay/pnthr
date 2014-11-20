package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"os"
	"strings"
)

type Instance struct {
	Identifier string
	Secret     string
	Password   string
}

/**
 * In the open command window set the following for Heroku:
 * heroku config:set MONGOHQ_URL=mongodb://user:pass@somesite.com:10027
 */
var URI string = os.Getenv("MONGO_URL")
var DBName string = os.Getenv("MONGO_DB")

func main() {
	if URI == "" {
		fmt.Println("no connection string provided, using localhost")
		URI = "localhost"
	}

	if DBName == "" {
		fmt.Println("no database name provided, bailing!")
		os.Exit(1)
	}

	/**
	 * POST /
	 *
	 * All requests will come through the root, via post
	 * Each request should have an app id and payload that has been encrypted with the app secret
	 * We want to take this payload, decrypt it with the app secret
	 * Once we have the raw payload, we encrypt first with the app password
	 * Secondly with the encrypt with the app secret, for transport back to the requestor
	 */
	http.HandleFunc("/", root)
	log.Println("Listening for connections...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func root(w http.ResponseWriter, r *http.Request) {
	/**
	 * Get the custom pnthr http header, which is our app ID
	 */
	id := r.Header.Get("pnthr")

	/**
	 * MongoDB Setup
	 *
	 * We are only connecting to a localhost for now
	 * In the future we'll want to support multiple nodes which just means we need to pass
	 * a comma separated list in the Dial() method
	 */
	session, err := mgo.Dial(URI)
	if err != nil {
		fmt.Printf("Can't connect to mongo, go error %v\n", err)
		panic(err)
	}
	defer session.Close()

	db := session.DB(DBName).C("instances")

	/**
	 * Failure: No App Key was passed
	 *
	 * If we don't have an api key in the header, then we can't fulfill this request
	 */
	if len(id) == 0 {
		Responder(w, r, 412, "Expected to find the 'pnthr' request header with your app id as the value, but none was found")
		return
	}

	/**
	 * Retrieve the application based on the passed App ID
	 * If none is found we need to change the response status code
	 */
	instance := Instance{}

	err = db.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&instance)

	if err != nil {
		Responder(w, r, 410, "That application ID is either invalid or was deleted")
		return
	}

	/**
	 * All has gone well with the request and app lookup
	 * Let's decrypt the payload and then re-encrypt it
	 */
	payloadRaw, err := ioutil.ReadAll(r.Body)

	if err != nil || string(payloadRaw) == "" {
		Responder(w, r, 422, "The request body was empty or unprocessable")
		return
	}

	/**
	* Payload comes in as <base64 payload>-<initialization vector>
	* therefore we'll split on the dash (since that's a safe char not part of the base64 map)
	* we'll need to verify that the split worked and that each piece is valid
	 */
	payloadParsed := strings.Split(string(payloadRaw), "-")
	payload := payloadParsed[0]
	iv := []byte(payloadParsed[1])[:aes.BlockSize]

	/**
	* Decrypt Transport for the insecure payload
	 */
	decoded := Base64Decode(payload)
	decrypted := make([]byte, len(string(decoded)))
	err = DecryptAES(decrypted, decoded, []byte(instance.Secret), iv)
	if err != nil {
		panic(err)
	}

	/**
	* Re-incrypt the insecure payload with the password
	 */
	encrypted := make([]byte, len(string(decrypted)))
	err = EncryptAES(encrypted, decrypted, []byte(instance.Password), iv)
	if err != nil {
		panic(err)
	}

	/**
	* Create the transport layer
	 */
	transport := make([]byte, len(string(encrypted)))
	err = EncryptAES(transport, encrypted, []byte(instance.Secret), iv)
	if err != nil {
		panic(err)
	}

	/**
	* Decryption test of transpart layer
	 */
	decryptFirst := make([]byte, len(string(transport)))
	err = DecryptAES(decryptFirst, transport, []byte(instance.Secret), iv)
	if err != nil {
		panic(err)
	}

	/**
	* Decryption test of the encryption layer
	 */
	decryptSecond := make([]byte, len(string(decryptFirst)))
	err = DecryptAES(decryptSecond, decryptFirst, []byte(instance.Password), iv)
	if err != nil {
		panic(err)
	}

	Responder(w, r, 200, string(Base64Encode(transport))+string('-')+string(iv))
}

/**
 * HTTP Erorr Handler
 *
 * Set the http status code and provide an error message
 */
func Responder(w http.ResponseWriter, r *http.Request, status int, message string) {
	w.WriteHeader(status)
	log.Println(message)
	fmt.Fprint(w, message)
}

func Base64Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func Base64Decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func EncryptAES(dst, src, key, iv []byte) error {
	aesBlockEncryptor, err := aes.NewCipher([]byte(key))
	if err != nil {
		return err
	}
	aesEncrypter := cipher.NewCFBEncrypter(aesBlockEncryptor, iv)
	aesEncrypter.XORKeyStream(dst, src)
	return nil
}

func DecryptAES(dst, src, key, iv []byte) error {
	aesBlockEncryptor, err := aes.NewCipher([]byte(key))
	if err != nil {
		return err
	}
	aesEncrypter := cipher.NewCFBEncrypter(aesBlockEncryptor, iv)
	aesEncrypter.XORKeyStream(dst, src)
	return nil
}
