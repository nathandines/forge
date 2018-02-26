package stacklib

import (
	"log"
	"math/rand"
	"regexp"
	"testing"
	"time"
	"unicode/utf8"
)

func random(rng *rand.Rand, min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func generateRandomNames(rng *rand.Rand) string {
	n := random(rng, 1, 128)
	runes := make([]rune, n)
	for i := 0; i < n; i++ {
		runes[i] = rune(rng.Intn(utf8.MaxRune))
	}
	return string(runes)
}

func TestVerifyStackNameLengthInvalid(t *testing.T) {
	stackNames := []string{
		"",
	}

	// Generate 1000 random length strings filled with 'a'
	seed := time.Now().UTC().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	for i := 0; i < 1000; i++ {
		n := random(rng, 129, 1129)
		bytes := make([]byte, n)
		for i := 0; i < n; i++ {
			bytes[i] = 'a'
		}
		stackNames = append(stackNames, string(bytes))
	}

	for _, s := range stackNames {
		if err := verifyStackName(s); err == nil {
			t.Errorf("verifyStackName(%s) - want: error", s)
		}
	}
}

func TestVerifyStackNameRegexValid(t *testing.T) {
	stackNames := []string{
		"test-stack-name",
		"Test-Stack-Name",
		"test-stack-name1",
		"Test-Stack-Name1",
		"test-stack1-name",
		"Test-Stack1-Name",
		"test1-stack-name",
		"Test1-Stack-Name",
		"1test-stack-name",
		"1Test-Stack-Name",
		"test-1stack-name",
		"Test-1Stack-Name",
		"test-stack-1name",
		"Test-Stack-1Name",
		"test-stack-n4me",
		"Test-Stack-N4me",
		"test-st4ck-name",
		"Test-St4ck-Name",
		"t3st-stack-name",
		"T3st-Stack-Name",
		"TESTSTACKNAME",
		"teststackname",
		"TeStStAcKnAmE",
		"TeststacknamE",
		"tESTSTACKNAMe",
	}
	for _, s := range stackNames {
		if err := verifyStackName(s); err != nil {
			t.Errorf("verifyStackName(%s) - want: no error", s)
		}
	}
}

func TestVerifyStackNameRegexInvalid(t *testing.T) {
	stackNames := []string{
		"-teststackname",
		"teststackname-",
		"-",
		"--",
	}

	// Generate strings with random data in them which don't constitute a valid
	// stack name
	seed := time.Now().UTC().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	for i := 0; i < 1000; i++ {
		r := generateRandomNames(rng)
		// Don't put a valid Stack Name into the list of values
		m, err := regexp.MatchString("^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]$", r)
		if err != nil {
			log.Fatal(err)
		}
		if !m {
			stackNames = append(stackNames, r)
		}
	}

	for _, s := range stackNames {
		if err := verifyStackName(s); err == nil {
			t.Errorf("verifyStackName(%s) - want: error", s)
		}
	}
}
