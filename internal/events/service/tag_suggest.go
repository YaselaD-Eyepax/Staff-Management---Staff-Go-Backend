package service

import "strings"

func SuggestTags(title, summary, body string) []string {
    text := strings.ToLower(title + " " + summary + " " + body)
    tags := make([]string, 0)

    // Simple rule-based detection
    if containsAny(text, []string{"maintenance", "downtime", "upgrade"}) {
        tags = append(tags, "maintenance")
    }
    if containsAny(text, []string{"policy", "guideline", "rule"}) {
        tags = append(tags, "policy")
    }
    if containsAny(text, []string{"meeting", "sync", "discussion"}) {
        tags = append(tags, "meeting")
    }
    if containsAny(text, []string{"deadline", "due", "submit"}) {
        tags = append(tags, "deadline")
    }
    if containsAny(text, []string{"emergency", "urgent", "immediately"}) {
        tags = append(tags, "urgent")
    }
    if containsAny(text, []string{"holiday", "vacation"}) {
        tags = append(tags, "holiday")
    }
    if containsAny(text, []string{"system", "server", "deploy"}) {
        tags = append(tags, "system")
    }

    return tags
}

func containsAny(text string, words []string) bool {
    for _, w := range words {
        if strings.Contains(text, w) {
            return true
        }
    }
    return false
}
