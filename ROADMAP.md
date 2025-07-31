# Karu Roadmap

## Features

### UX
- [x] Add download functionality to save episodes for offline viewing
- [x] Implement watch history/resume feature to track progress
- [ ] Add favorites/watchlist management
- [x] Create a recently watched menu for quick access
- [x] Add quality selection (720p, 1080p, etc.) when multiple sources available
- [x] Implement auto-play next episode option
- [x] Add video player controls with TUI interface
- [x] Add loading messages throughout the application
- [ ] Add subtitle selection and downloading

### Config
- [x] Add config file support for default player, quality preferences, download directory
- [x] Allow custom player commands and arguments
- [ ] Add theme/color customization for the UI
- [ ] Implement custom keybindings

### Search
- [ ] Add genre-based browsing and filtering
- [ ] Implement advanced search with filters (year, status, rating)
- [x] Add "trending" or "popular" anime discovery
- [x] Create search history with quick access
- [ ] Add fuzzy search improvements

### Smaller UX stuff
- [x] Add concurrent episode loading for faster browsing
- [x] Implement caching for search results and episode lists
- [x] Add progress bars for downloads and loading
- [x] Hide video player logs for cleaner interface
- [ ] Create update checker for new Karu versions

## Bug Fixes

### Error Handling
- [ ] Better error messages when video player is not found
- [ ] Handle network timeouts more gracefully
- [ ] Add retry logic for failed scraping attempts
- [ ] Validate URLs before attempting to play

### Player Detection
- [ ] Improve player detection logic (check PATH, common install locations)
- [ ] Add fallback player options when primary choice fails
- [ ] Handle missing player dependencies better

### Cross-Platform Bugs
- [ ] Fix Windows-specific path handling
- [ ] Improve macOS player detection beyond just iina
- [ ] Add proper signal handling for clean exits

### UI/UX Fixes
- [ ] Handle empty search results more elegantly
- [ ] Fix potential crashes when selecting invalid episodes
- [ ] Improve keyboard navigation and selection
- [ ] Add proper exit handling in interactive menus

### Scraping Reliability
- [ ] Add user-agent rotation to avoid blocking
- [ ] Implement source failover when primary source fails
- [ ] Handle different video source formats
- [ ] Add rate limiting to prevent IP blocking
