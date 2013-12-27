/**
 * Author: Andrey Kolchenko <andrey@kolchenko.me>
 * Date: 25.11.13
 */
package session

type SortedSessions struct {
	Session *Session
	Weight  int
}

type CapabilitiesSorter struct {
	sortedCapabilities  []*SortedSessions
	desiredCapabilities *Capabilities
}

func NewSorter(desiredCapabilities Capabilities) *CapabilitiesSorter {
	var cs *CapabilitiesSorter = new(CapabilitiesSorter)
	cs.desiredCapabilities = &desiredCapabilities
	return cs
}

func (cs *CapabilitiesSorter) Add(session *Session) {
	if weight, suitable := cs.weight(session); suitable {
		forSort := SortedSessions{session, weight}
		cs.sortedCapabilities = append(cs.sortedCapabilities, &forSort)
	}
}

func (cs *CapabilitiesSorter) GetIterator() []*SortedSessions {
	return cs.sortedCapabilities
}

func (cs *CapabilitiesSorter) weight(session *Session) (int, bool) {
	dc := cs.desiredCapabilities
	var scores int
	capabilities := session.Capabilities
	if dc.BrowserName == capabilities.BrowserName {
		if dc.Platform.Any() {
			if capabilities.Platform.Any() {
				scores += 8
			} else {
				scores += 6
			}
		} else {
			if capabilities.Platform.Any() {
				scores += 7
			} else if capabilities.Platform == dc.Platform {
				scores += 4
			} else {
				return 0, false
			}
		}
		if dc.Version.Any() {
			if capabilities.Version.Any() {
				scores += 4
			} else {
				scores += 2
			}
		} else {
			if capabilities.Version.Any() {
				scores += 3
			} else if capabilities.Version == dc.Version {
				scores += 1
			} else {
				return 0, false
			}
		}
		scores += session.GetWeight()
		return scores, true
	}
	return 0, false
}

func (cs *CapabilitiesSorter) Len() int {
	return len(cs.sortedCapabilities)
}

func (cs *CapabilitiesSorter) Less(i, j int) bool {
	return cs.sortedCapabilities[i].Weight < cs.sortedCapabilities[j].Weight
}

func (cs *CapabilitiesSorter) Swap(i, j int) {
	cs.sortedCapabilities[i], cs.sortedCapabilities[j] = cs.sortedCapabilities[j], cs.sortedCapabilities[i]
}
