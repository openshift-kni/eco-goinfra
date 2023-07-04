package nad

// IPAMStatic returns static ipam type.
func IPAMStatic() *IPAM {
	return &IPAM{Type: "static"}
}

// IPAMWhereAbout returns WhereAbout ipam type.
func IPAMWhereAbout(ipRange, gateway string) *IPAM {
	ipam := &IPAM{Type: "whereabouts", IPRanges: []IPRanges{{Range: ipRange, Gateway: gateway}}}

	if ipRange == "" {
		return nil
	}

	if gateway == "" {
		return nil
	}

	return ipam
}

// WhereAboutAppendRange returns WhereAbout ipam type with additional address range.
func WhereAboutAppendRange(ipam *IPAM, ipRange, gateway string) *IPAM {
	ipam.IPRanges = append(ipam.IPRanges, IPRanges{Range: ipRange, Gateway: gateway})

	if ipRange == "" {
		return nil
	}

	if gateway == "" {
		return nil
	}

	return ipam
}
