# Recent Changes to f6n

## Color Theme Update ✨

The application now features a new, modern color scheme:

- **Background**: Dark grey (#1c1c1c) instead of pure black for better readability
- **ASCII Art**: Golden yellow (#FFD700) for the f6n logo
- **Command Keys**: Hot pink (#FF69B4) for keyboard shortcuts (e.g., `<enter>`, `<l>`)
- **Command Values**: Grey (#808080) for shortcut descriptions
- **Other elements**: Maintained cyan/teal accent colors for highlights

## UI Improvements 🎨

### 3-Column Keyboard Shortcuts Layout

Shortcuts are now displayed in an organized 3-column grid format:

```
<enter>: view details  │  <l>: logs          │  <e>: change env
<a>: api gateway       │  <c>: view code     │  <r>: refresh
<q>: quit
```

This layout:
- Makes better use of horizontal space
- Is easier to scan visually
- Groups related commands logically

## Command-Line Arguments 🚀

Added comprehensive CLI flags:

```bash
# Show version
f6n --version
f6n -v

# Show help
f6n --help
f6n -h

# Specify AWS region
f6n --region us-west-2

# Set environment
f6n --env prod

# Use AWS profile
f6n --profile production

# Set log level
f6n --log-level debug
```

## Testing

Build and test the changes:

```bash
# Build
make build

# Test version flag
./bin/f6n --version

# Test help
./bin/f6n --help

# Run the app
./bin/f6n
```

## Next Steps

Planned features ready to implement:
- [ ] CloudWatch Logs integration
- [ ] API Gateway endpoint discovery
- [ ] Function code viewer
- [ ] Environment switching (e key)
- [ ] Enhanced navigation and views

All infrastructure is in place to add these features incrementally.
