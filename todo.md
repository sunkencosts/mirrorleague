# Next
The sleeper route https://api.sleeper.app/v1/league/1182073403987832832/matchups/1 returns the player scores for the matchup. We need to return these in our matchup routes. Basically we want to know, for every week, what did every player in the matchup score? We also need to double check that the lineup feature works only players that were on the team at the time. Like, can I scroll back and override with a new addition to the team that wasn't there at the time?

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