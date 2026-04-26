package rankings

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{"Josh Allen", "josh allen"},
		{"C.J. Stroud", "cj stroud"},
		{"Patrick Mahomes II", "patrick mahomes"},
		{"Travis Kelce Jr", "travis kelce"},
		{"A.J. Brown", "aj brown"},
		{"Ja'Marr Chase", "jamarr chase"},
		{"Tre' Harris", "tre harris"},
		{"De’Zhaun Stribling", "dezhaun stribling"},
	}
	for _, tc := range cases {
		if got := normalize(tc.input); got != tc.want {
			t.Errorf("normalize(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestRarityFromPct(t *testing.T) {
	cases := []struct {
		pct  float64
		want string
	}{
		{0.01, "mythic"},
		{0.04, "orange"},
		{0.15, "purple"},
		{0.30, "blue"},
		{0.60, "green"},
		{0.90, "grey"},
	}
	for _, tc := range cases {
		if got := rarityFromPct(tc.pct); got != tc.want {
			t.Errorf("rarityFromPct(%v) = %q, want %q", tc.pct, got, tc.want)
		}
	}
}

func TestLoad_assignsRarity(t *testing.T) {
	header := "fp_page,page_type,ecr_type,player,id,pos,team,ecr,sd,best,worst,sportsdata_id,player_filename,yahoo_id,cbs_id,player_owned_avg,player_owned_espn,player_owned_yahoo,player_image_url,player_square_image_url,rank_delta,bye,mergename,scrape_date,tm"
	cols := make([]string, 25)
	for i := range cols {
		cols[i] = "NA"
	}
	cols[1] = "dynasty-wr"
	cols[5] = "WR"
	cols[22] = "Some Player"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, header)
		fmt.Fprintln(w, strings.Join(cols, ","))
	}))
	defer srv.Close()

	c := &Cache{cacheFile: filepath.Join(t.TempDir(), "rankings.json")}
	if err := c.Load(srv.URL); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, ok := c.Get("Some Player", "WR"); !ok {
		t.Error("Get: player not found after Load")
	}
}

func TestLoad_topPlayerAlwaysMythic(t *testing.T) {
	// 5 players: rank 1 = 20%, well outside the mythic % threshold.
	// The top player should still be mythic; rank 2 should fall through to %.
	header := "fp_page,page_type,ecr_type,player,id,pos,team,ecr,sd,best,worst,sportsdata_id,player_filename,yahoo_id,cbs_id,player_owned_avg,player_owned_espn,player_owned_yahoo,player_image_url,player_square_image_url,rank_delta,bye,mergename,scrape_date,tm"
	names := []string{"Top QB", "Second QB", "Third QB", "Fourth QB", "Fifth QB"}
	lines := []string{header}
	for _, name := range names {
		cols := make([]string, 25)
		for i := range cols {
			cols[i] = "NA"
		}
		cols[1] = "dynasty-qb"
		cols[5] = "QB"
		cols[22] = name
		lines = append(lines, strings.Join(cols, ","))
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, strings.Join(lines, "\n"))
	}))
	defer srv.Close()

	c := &Cache{cacheFile: filepath.Join(t.TempDir(), "rankings.json")}
	if err := c.Load(srv.URL); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got, _ := c.Get("Top QB", "QB"); got != "mythic" {
		t.Errorf("top player: got %q, want mythic", got)
	}
	if got, _ := c.Get("Second QB", "QB"); got == "mythic" {
		t.Errorf("second player should not be mythic by position rule")
	}
}
