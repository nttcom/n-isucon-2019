package utils

import (
	"encoding/base32"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

// SetStatus set response code
func SetStatus(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
}

// JSONToMap converts JSON response to Map
func JSONToMap(w http.ResponseWriter, r *http.Request) map[string]interface{} {

	length, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		log.Printf("Content-Length NG\n")
		SetStatus(w, 400)
		return nil
	}

	log.Printf("Content-Length OK\n")

	body := make([]byte, length)
	length, err = r.Body.Read(body)
	if err != nil && err != io.EOF {
		SetStatus(w, 400)
		return nil
	}

	//json to map
	var jsonReq map[string]interface{}
	err = json.Unmarshal(body[:length], &jsonReq)
	if err != nil {
		// If the JSON is not convertible.
		SetStatus(w, 400)
		return nil
	}
	return jsonReq
}

// GetSalt generates random string.
func GetSalt() string {
	b := make([]byte, 64)
	_, err := rand.Read(b)
	if err != nil {
		panic("Salt generation error")
	}

	return strings.TrimRight(base32.StdEncoding.EncodeToString(b), "=")
}

// GetPasswordHash generates password hash.
func GetPasswordHash(salt string, password string) string {
	result := password + salt

	for i := 0; i < 1000; i++ {
		out, err := exec.Command("sh", "-c", "echo -n '"+result+"' | openssl sha256").Output()
		if err != nil {
			panic("Failed to generate the hash.")
		}

		outputRaw := string(out)
		output := strings.Split(outputRaw, " ")
		result = strings.TrimRight(output[1], "\n")

	}
	return result
}

// ContentTypeEqualJSON checks contentType='application/json'.
func ContentTypeEqualJSON(contentType string) bool {
	return strings.Contains(contentType, "application/json")
}
