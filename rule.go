package RateLimiter

// The Rule of one uri. the same user can just access this uri 'Limit' times in 'Durantion' time
type Rule struct {
	Limit    int32
	Duration int32 // Units per second
}

type Rules map[string][]*Rule

func (rules Rules) AddRule(path string, rule *Rule) {
	if _, ok := rules[path]; !ok {
		rules[path] = make([]*Rule, 0)
	}

	rules[path] = append(rules[path], rule)
}

func NewRules() Rules {
	return make(Rules)
}
