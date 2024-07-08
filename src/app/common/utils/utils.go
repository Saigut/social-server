package utils

import (
    "crypto/md5"
    "encoding/hex"
)

func CalPassHash(password string) string {
    combined := "ss" + password
    hash := md5.New()
    hash.Write([]byte(combined))
    hashInBytes := hash.Sum(nil)
    return hex.EncodeToString(hashInBytes)
}
