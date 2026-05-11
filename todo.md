# Next
Need to updat the way the score is display for:
- each player. Have something more in the center with and an easy way to show +-
- At the top for win/lose have +- as well.
- it's always waiting for GET http://localhost:5173/api/lineups?user_id=2240d6f1-d529-42ec-b36b-1df8054a284f&league_id=1182073403987832832&week_number=1 to return before updating which is slow because we already have all the scores per player when they load.

# FE
- Show scores red/green after player starts.
- Use sleeper historical data to test.
- search for a player
- figure out a better way to rank teams

# Pages
- Some type of stats page for the league? 
- Some type of leaderboard for the entire site. Average closest to max points for? 

# Database



Entire point is to submit a better lineup and see if you beat the owner.
- Testing should be easy. Just use 2025 data and simulate.
- Incentive for teams who suck like Kevin, it's an incentive to stay involved in the league. Maybe the league says that they get a perk for setting the best lineup?
- How can I claim a team and ask "who should I start" questions? Or is that implicit by allowing people to set any lineup.



# Bugs
- Bad league ID is not handled on "Connect league"

# Auth
- figure out how testing works. I don't want to have to auth every time.
- Dev login: hit GET /api/dev/login (only works when APP_ENV=development). Sets auth_token cookie, redirects to frontend. Accepts optional ?user_id=&email=&username= query params.
- Logout clears the cookie client-side but the JWT stays valid server-side until it expires (30 days). This is accepted — stateless JWTs can't be revoked without a server-side blocklist. If token revocation becomes important, add a short expiry + refresh token or a Redis blocklist.