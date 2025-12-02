package update

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
)

/*

Generate private key:
openssl genrsa 4096 > sign_test_key.txt
openssl rsa -in sign_test_key.txt -pubout > sign_test_pubkey.txt

Sign:
openssl dgst -sha256 -sign sign_test_key.txt sign_test_binary > sign_test_sign

Verify:
openssl dgst -sha256 -verify sign_test_pubkey.txt -signature sign_test_sign sign_test_binary
*/

const EXAMPLE_PRIVATE_KEY = `
-----BEGIN PRIVATE KEY-----
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
RBpJV4CczHTR88z9MDAij4ruW4AMm3g=
-----END PRIVATE KEY-----
`

const EXAMPLE_PUBLIC_KEY = `
-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAq93R3k/JO/DJD0J4Sbmd
TMPc5uWS5oR+G6FfXre8DFDuXQmZTg0GKUZaImeZJoF0PkQ6YChUAQwM0NH5sYiu
iiyWwWIjZduKrd7dNslXOFhdcH57EiWEY0niX8VJF7LzVa+HIZPdMMw4NEGsch/G
AsBgSzeck4Hq+3PL9I+yrsHqeoQj50VYGLZQvqc43vMs/YhHNI7lvwXu6Zp0uSeP
iQU1OAoNfbYJ+56iuIPWKmsniAd0rmHrqkgdezo1caA2grqP/JkQ75mQ67fjiE85
/0haYKfhbaqUX3oooi1xYdNUkr/wjubf8J9Sksk58qgUXl7uVYQ4OtIHiTHM+UnE
ISrAomI5flXX05r1c4K0eba4/IdDYUZ1yiEkvQ76gAPAZb+HSasLJB2K0xPxAMnT
GQwwc4/0rJ8QxyQGHX4vPtK+eeAQNyNwH43uQXpovzLKeWIV7TlKjyFoMIwhHyxw
rzKHIGNgGQUcoeuP2npRkptBjtkipCwxLbA0Y1aW0d5WhIOBXxZxkbS3v/2oNpcO
hH4BRUYI0rQ8194ra/L4SGT+Qp2a6SNfu3fLQzBEBFgG9lyN7ZpqV74yJc4eRV6f
Gh431TfNxUBxMwRwK4EjyTJfxeOQ9SXyCNCW1xJP7uO/rYxVTixw8DaXwwqZuEV5
n0t+QDTU3QWmYclpOluxo4MCAwEAAQ==
-----END PUBLIC KEY-----
`

// Compares our signature to the signature produced by openssl
func TestSign(t *testing.T) {
	binaryBytes, err := os.ReadFile("sign_test_binary")
	if err != nil {
		t.Fatal(err)
	}

	s, err := Sign(binaryBytes, EXAMPLE_PRIVATE_KEY)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.Buffer{}

	cmd := exec.Command("openssl", "dgst", "-sha256", "-sign", "sign_test_key.txt", "sign_test_binary")
	cmd.Stdout = &b
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(s, b.Bytes()) {
		t.Fatal("not equal")
	}
}

func TestVerify(t *testing.T) {
	binaryBytes, err := os.ReadFile("sign_test_binary")
	if err != nil {
		t.Fatal(err)
	}

	s, err := Sign(binaryBytes, EXAMPLE_PRIVATE_KEY)
	if err != nil {
		t.Fatal(err)
	}

	err = Verify(binaryBytes, EXAMPLE_PUBLIC_KEY, s)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVerifyBadBinary(t *testing.T) {
	binaryBytes, err := os.ReadFile("sign_test_binary")
	if err != nil {
		t.Fatal(err)
	}

	s, err := Sign(binaryBytes, EXAMPLE_PRIVATE_KEY)
	if err != nil {
		t.Fatal(err)
	}

	binaryBytes[0] = '6'

	err = Verify(binaryBytes, EXAMPLE_PUBLIC_KEY, s)
	if err == nil {
		t.Fatal()
	}
}

func TestVerifyBadBinary2(t *testing.T) {
	binaryBytes, err := os.ReadFile("sign_test_binary")
	if err != nil {
		t.Fatal(err)
	}

	s, err := Sign(binaryBytes, EXAMPLE_PRIVATE_KEY)
	if err != nil {
		t.Fatal(err)
	}

	binaryBytes = append(binaryBytes, 0x6)

	err = Verify(binaryBytes, EXAMPLE_PUBLIC_KEY, s)
	if err == nil {
		t.Fatal()
	}
}

func TestVerifyBadSign(t *testing.T) {
	binaryBytes, err := os.ReadFile("sign_test_binary")
	if err != nil {
		t.Fatal(err)
	}

	s, err := Sign(binaryBytes, EXAMPLE_PRIVATE_KEY)
	if err != nil {
		t.Fatal(err)
	}

	s[0] = s[0] + 1

	err = Verify(binaryBytes, EXAMPLE_PUBLIC_KEY, s)
	t.Log(err)
	if err == nil {
		t.Fatal()
	}
}

func TestVerifyBadSign2(t *testing.T) {
	binaryBytes, err := os.ReadFile("sign_test_binary")
	if err != nil {
		t.Fatal(err)
	}

	s, err := Sign(binaryBytes, EXAMPLE_PRIVATE_KEY)
	if err != nil {
		t.Fatal(err)
	}

	s = append(s, 0x6)

	err = Verify(binaryBytes, EXAMPLE_PUBLIC_KEY, s)
	t.Log(err)
	if err == nil {
		t.Fatal()
	}
}
