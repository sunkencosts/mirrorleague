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
	}
	for _, tc := range cases {
		if got := normalize(tc.input); got != tc.want {
			t.Errorf("normalize(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestLoad_rarityTiers(t *testing.T) {
	// 20 QBs lets us hit every tier cleanly:
	//   rank 1  →  5% = orange
	//   rank 4  → 20% = purple
	//   rank 9  → 45% = blue
	//   rank 15 → 75% = green
	//   rank 20 → 100% = grey
	names := []string{
		"Player Orange",
		"Player Purple A", "Player Purple B", "Player Purple C",
		"Player Blue A", "Player Blue B", "Player Blue C", "Player Blue D", "Player Blue E",
		"Player Green A", "Player Green B", "Player Green C", "Player Green D", "Player Green E", "Player Green F",
		"Player Grey A", "Player Grey B", "Player Grey C", "Player Grey D", "Player Grey E",
	}

	header := "fp_page,page_type,ecr_type,player,id,pos,team,ecr,sd,best,worst,sportsdata_id,player_filename,yahoo_id,cbs_id,player_owned_avg,player_owned_espn,player_owned_yahoo,player_image_url,player_square_image_url,rank_delta,bye,mergename,scrape_date,tm"
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

	cases := []struct {
		name, pos, want string
	}{
		{"Player Orange", "QB", "orange"},
		{"Player Purple A", "QB", "purple"},
		{"Player Blue A", "QB", "blue"},
		{"Player Green A", "QB", "green"},
		{"Player Grey A", "QB", "grey"},
	}
	for _, tc := range cases {
		got, ok := c.Get(tc.name, tc.pos)
		if !ok {
			t.Errorf("Get(%q, %q): not found, want %q", tc.name, tc.pos, tc.want)
			continue
		}
		if got != tc.want {
			t.Errorf("Get(%q, %q) = %q, want %q", tc.name, tc.pos, got, tc.want)
		}
	}
}
