package yokai

// extraAliases is populated from the embedded catalog at init.
var extraAliases = map[string]string{}

// ExtraAliases returns a copy of the extra spelling map (for catalog export).
func ExtraAliases() map[string]string {
	out := make(map[string]string, len(extraAliases))
	for k, v := range extraAliases {
		out[k] = v
	}
	return out
}
