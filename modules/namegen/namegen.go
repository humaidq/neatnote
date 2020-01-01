package namegen

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var (
	// ADJ is a list of adjectives used in the name generator.
	ADJ = []string{"adaptable", "adventurous", "ambitious", "amusing",
		"agreeable", "brave", "bright", "calm", "charming", "considerate",
		"courageous", "creative", "decisive", "diligent", "diplomatic",
		"discreet", "dynamic", "enthusiastic", "exuberant", "faithful",
		"fearless", "friendly", "fearless", "funny", "generous", "gentle",
		"gregarious", "helpful", "honest", "humorous", "imaginative",
		"impartial", "idependent", "intellectual", "kind", "loving", "loyal",
		"neat", "nice", "passionate", "persistent", "polite", "powerful",
		"quiet", "rational", "reliable", "romantic", "thoughtful", "tidy"}

	// NOUN is a list of nouns used in the name generator.
	NOUN = []string{"aardvark", "albatross", "alligator", "alpaca", "ant",
		"antelope", "badger", "bat", "bear", "bee", "bird", "butterfly",
		"camel", "caibou", "cassowary", "cat", "chicken", "chinchilla",
		"chough", "coati", "cobra", "cod", "crab", "crow", "cuckoo", "deer",
		"dolphin", "dragonfly", "duck", "eagle", "eel", "emu", "falcon",
		"ferret", "finch", "frog", "gecko", "gnu", "bagot", "kiko", "serow",
		"goose", "horse", "hyena", "jellyfish", "kangaroo", "koala", "kudu",
		"lapwing", "lion", "lynx", "mink", "mongoose", "oryx", "otter", "owl",
		"oyster", "unicorn", "pelican", "pony", "turtle", "weasel", "wren",
		"zebra"}
)

var rnd *rand.Rand

func init() {
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// GetName returns a randomly generated name which is formatted.
func GetName() (name string) {
	name = fmt.Sprintf("%s %s", ADJ[rnd.Intn(len(ADJ))],
		NOUN[rnd.Intn(len(NOUN))])
	name = strings.Title(name)
	return
}
