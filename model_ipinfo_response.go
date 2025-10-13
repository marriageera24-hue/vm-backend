package main

type IPInfoIPAddressResponse struct {
	IP  string `json:"ip"`
	ASN struct {
		ASN    string `json:"asn"`
		Name   string `json:"name"`
		Domain string `json:"domain"`
		// Route  string `json:"route"`
		Type string `json:"type"`
	} `json:"asn"`
	Company struct {
		Type string `json:"type"`
	} `json:"company"`
}
