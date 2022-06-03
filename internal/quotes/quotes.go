package quotes

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/robotomize/powwy/pkg/hashcash"
)

var quotes = []string{
	`“Voice is not just the sound that comes from your throat, but the feelings that come from your words.”
― Jennifer Donnelly, A Northern Light`, `“Quit being so hard on yourself. We are what we are; we love what we love. We don't need to justify it to anyone... not even to ourselves.”
― Scott Lynch, The Republic of Thieves`, `“I like to listen. I have learned a great deal from listening carefully. Most people never listen.”
― Ernest Hemingway`, `“There isn't any questioning the fact that some people enter your life, at the exact point of need, want or desire - it's sometimes a coincendence and most times fate, but whatever it is, I am certain it came to make me smile.”
― Nikki Rowe`, `“look for a long time at what pleases you, and longer still at what pains you...”
― Colette`, `“Socrates: Have you noticed on our journey how often the citizens of this new land remind each other it is a free country?
Plato: I have, and think it odd they do this.
Socrates: How so, Plato?
Plato: It is like reminding a baker he is a baker, or a sculptor he is a
sculptor.
Socrates: You mean to say if someone is convinced of their trade, they have
no need to be reminded.
Plato: That is correct.
Socrates: I agree. If these citizens were convinced of their freedom, they would not need reminders.”
― E.A. Bucchianeri, Brushstrokes of a Gadfly,`, `“The worst part of being okay is that okay is far from happy.”
― Anna Todd`, `“Healing is more about accepting the pain and finding a way to peacefully co-exist with it. In the sea of life, pain is a tide that will ebb and weave, continually.

We need to learn how to let it wash over us, without drowning in it. Our life doesn't have to end where the pain begins, but rather, it is where we start to mend.”
― Jaeda DeWalt`, `“Take care of your words and the words will take care of you.”
― Amit Ray`, `“The only way to get what you want is to make them more afraid of you than they are of each other.”
― Cinda Williams Chima, The Crimson Crown`, `“There is something incredibly beautiful about a woman, who knows herself, she can't break, she just falls but in every fall she rises, past who she was before.”
― Nikki Rowe`, `“Success in life is not for those who run fast, but for those who keep running and always on the move.”
― Bangambiki Habyarimana, Pearls Of Eternity`, `“You push the TRUTH off a cliff, but it will always fly. You can submerge the TRUTH under water, but it will not drown. You can place the TRUTH in the fire, but it will survive. You can bury the TRUTH beneath the ground, but it will arise. TRUTH always prevails!”
― Amaka Imani Nkosazana, Heart Crush`, `“Oh darling, your only too wild, to those whom are to tame, don't let opinions change you.”
― Nikki Rowe`, `“Your life is a movie. You are the main character. You say your scripts and act to your lines. Of course you do your lines in each scene. There is a hidden camera and a director who you can ask for help anytime up above.”
― Happy Positivity`,
}

func NewQuotes(config Config) *Quotes {
	return &Quotes{config: config}
}

type Quotes struct {
	config Config
}

func (a *Quotes) MakeChallenge(subject string) (hashcash.Header, error) {
	expired := time.Now().Add(a.config.HashCashExpiredDuration).Unix()
	header, err := hashcash.Default(subject, a.config.HashCashDifficult, expired)
	if err != nil {
		return hashcash.Header{}, fmt.Errorf("hashcash.Default: %w", err)
	}

	return header, nil
}

func (a *Quotes) GetResource() (string, error) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint
	idx := rnd.Intn(len(quotes) - 1)

	return quotes[idx], nil
}
