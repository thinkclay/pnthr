package main

import (
  "fmt"
  // "crypto/aes"
  // "crypto/cipher"
  // "encoding/base64"
  // "fmt"
  // "github.com/codegangsta/martini"
  // "io/ioutil"
  "labix.org/v2/mgo"
  // "labix.org/v2/mgo/bson"
  // "net/http"
  // "os"
  // "strings"
  "testing"
)

func TestMongoPresence(t *testing.T) {
  session, err := mgo.Dial("mongodb://localhost:27017")
  if err != nil {
    t.Error("Can't connect to mongo", err)
  }
  defer session.Close()
}

func TestEncryption(t *t.testing.T) {
  // id: 538362a63832640002020000
  // secret: 8d1067143a608920a56f4d4a7c6e3d4b
  password := "22e5ab5743ea52caf34abcc02c0f161d"
  iv := "538362a638326400"

  /**
   * Re-incrypt the insecure payload with the password
   */
  encrypted := make([]byte, len(string("this is a test")))
  err = EncryptAES(encrypted, decrypted, []byte(password), iv)
  if err != nil {
    panic(err)
  }


  Base64Encode(encrypted) == "Xwt+WH8rcvyxw6t28LA=-538362a638326400"
}
