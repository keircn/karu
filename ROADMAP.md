# Karu Roadmap

## Features

### UX
- [ ] Add download functionality to save episodes for offline viewing
- [ ] Implement watch history/resume feature to track progress
- [ ] Add favorites/watchlist management
- [ ] Create a recently watched menu for quick access
- [ ] Add quality selection (720p, 1080p, etc.) when multiple sources available
- [ ] Implement auto-play next episode option
- [ ] Add subtitle selection and downloading

### Config
- [ ] Add config file support for default player, quality preferences, download directory
- [ ] Allow custom player commands and arguments
- [ ] Add theme/color customization for the UI
- [ ] Implement custom keybindings

### Search
- [ ] Add genre-based browsing and filtering
- [ ] Implement advanced search with filters (year, status, rating)
- [ ] Add "trending" or "popular" anime discovery
- [ ] Create search history with quick access
- [ ] Add fuzzy search improvements

### Smaller UX stuff
- [ ] Add concurrent episode loading for faster browsing
- [ ] Implement caching for search results and episode lists
- [ ] Add progress bars for downloads and loading
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
