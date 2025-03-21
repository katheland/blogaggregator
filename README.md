# blogaggregator
A blog aggregator CLI for Boot.dev

# Setup
You will need Go and Postgresql installed in order to run this program.

Go: `curl.exe https://webi.ms/golang | powershell`

Postgresql: `sudo apt install postgresql postgresql-contrib`

Once you have Postgresql installed, you'll want to start it in the background (`sudo service postgresql start`).

If you want the generated database files, you might also want to install SQLC (`go install github.om/sqlc-dev/sqlc/cmd/sqlc@latest`) and run `sqlc generate` from the root of the project.

You should be able to run it using `./gator <...>` (see below), but if you want to install it, use `go install` from the root of the project.  Then you should be able to call `gator <...>` from anywhere.

You'll also need to create a .gatorconfig.json file in your home directory.  It should contain the following:

```
{
  "db_url": "postgres://postgres:@localhost:5432/gator"
}
```

(Modify it if you decide to make alterations to your Postgresql server.)

# Running Gator
- `gator register <username>` - Registers a user and switches to them
- `gator login <username>` - Switches to the user if they exist
- `gator users` - Lists all of the users
- `gator addfeed <title> <url>` - Adds a feed, if it hasn't been added yet
- `gator feeds` - Lists all of the feeds
- `gator follow <url>` - Follow a feed that someone else added
- `gator unfollow <url>` - Unfollow a feed
- `gator following` - Lists all of the feeds you're following
- `gator agg <duration>` - Runs perpetually until stopped.  After each duration (ex. 5s, 1m, 1h), retrieves and stores the latest posts from the followed feed that was last checked the longest time ago.
- `gator browse [limit (default 2)]` - Shows the latest posts from your feed.

# Ideas For The Future
- I mean, login could have actual credentials, for one.
- Filtering and/or paginating the browse could be useful.
- Web browser support of some kind?
