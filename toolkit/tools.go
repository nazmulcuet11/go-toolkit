package toolkit

import "math/rand"

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
		index := rand.Intn(len(r))
		s[i] = r[index]
	}
	return string(s)
}
