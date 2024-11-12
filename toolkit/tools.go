package toolkit

import "crypto/rand"

const randomCharacterSet = "abcdefghijklijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01223456789"

// Tools is the type to instantiate the module. Any variable of this type have the access to all
// the methods with the receiver *Tools.
type Tools struct {
}

// RandomString returns a string of random character of length n.
func (t *Tools) RandomString(n int) string {
	s := make([]rune, n)
	r := []rune(randomCharacterSet)
	for i := range s {
		p, _ := rand.Prime(rand.Reader, 64)
		index := p.Uint64() % uint64(len(r))
		s[i] = r[index]
	}
	return string(s)
}
