package importer

import "testing"

// TestING_KontoauszugNote: a counterparty's name longer than the column
// wraps mid-word onto the Valuta line; the note joins it back together
// and separates the Verwendungszweck with " · ".
func TestING_KontoauszugNote(t *testing.T) {
	text := "ING-DiBa AG\nKontoauszug 2026\nIBAN DE02 1203 0000 0000 2020 51\n" +
		"Buchung      Buchung / Verwendungszweck                                     Betrag (EUR)\n" +
		"Valuta\n\n" +
		"28.09.2026                 Gehalt/Rente Fraunhofer-Gesellschaft zur Forderu                 4.000,00\n" +
		"28.09.2026                 ng der angewandten Forschung einge\n" +
		"                           3440 Lohn/Gehalt 10010826/202609\n" +
		"                           Referenz: NOTPROVIDED\n" +
		"01.09.2026                 Lastschrifteinzug LBS SUED                                      -1.161,24\n" +
		"01.09.2026                 Bausparvertrag 7 038 222 803\n" +
		"                           Mandat: M123\n"
	txs, err := ING{}.Parse(text)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(txs) != 2 {
		t.Fatalf("len(txs) = %d, want 2", len(txs))
	}
	want := "Gehalt/Rente Fraunhofer-Gesellschaft zur Forderung der angewandten Forschung einge · 3440 Lohn/Gehalt 10010826/202609"
	if txs[0].Note != want {
		t.Errorf("note = %q\n   want %q", txs[0].Note, want)
	}
	if want := "Lastschrifteinzug LBS SUED · Bausparvertrag 7 038 222 803"; txs[1].Note != want {
		t.Errorf("note = %q, want %q", txs[1].Note, want)
	}
}
