# ⚠️ Known Limitation
- **Week traversal shows current rosters, not historical ones.** If a player was traded in week 10, going back to week 8 still shows them on their current team. Fix: use `/league/:leagueId/matchups/:week` (Sleeper API) instead of the current `/rosters` snapshot for past weeks. The rosters query needs to become week-aware.

# FE
- Show scores red/green after player starts.
- Use sleeper historical data to test.
- navigation - back to leagueId entry
- search for a player
- figure out a better way to rank teams

# Pages
- Some type of stats page for the league? 
- Some type of leaderboard for the entire site. Average closest to max points for? 

# Database
- User can see previously viewed leagueIds and set a nickname if they want.



Entire point is to submit a better lineup and see if you beat the owner.
- Testing should be easy. Just use 2025 data and simulate.
- Incentive for teams who suck like Kevin, it's an incentive to stay involved in the league. Maybe the league says that they get a perk for setting the best lineup?
- How can I claim a team and ask "who should I start" questions? Or is that implicit by allowing people to set any lineup.