# VHS Tape Script for Boa Terminal Demo
# Install VHS: brew install vhs
# Run: vhs docs/demo.tape

Output docs/img/demo.gif
Output docs/img/demo.mp4

Set FontSize 16
Set Width 1100
Set Height 680
Set Padding 20
Set Theme "Catppuccin Mocha"
Set WindowBar Colorful
Set TypingSpeed 35ms

# Launch Boa
Type "bo"
Sleep 700ms
Enter
Sleep 1800ms

# Show the main menu
Sleep 1500ms

# Open Pack
Down
Sleep 500ms
Enter
Sleep 1200ms

# Navigate file picker
Down
Sleep 500ms
Down
Sleep 500ms

# Select demo project
Enter
Sleep 1200ms

# Confirm compression
Enter
Sleep 3s

# Show compression result
Sleep 1800ms

# Return to main menu
Enter
Sleep 1000ms

# Clear screen
Type "clear"
Enter
Sleep 500ms

# Inspect generated archive
Type "bo list demo-project.zip"
Sleep 600ms
Enter
Sleep 3000ms

# Keep archive information visible
Sleep 1800ms

# Clear screen
Type "clear"
Enter
Sleep 500ms

# Benchmark compression
Type "bo bench ./demo-project"
Sleep 600ms
Enter
Sleep 5s

# Keep benchmark results visible
Sleep 2500ms

# Clear screen
Type "clear"
Enter
Sleep 500ms

# Demonstrate unpacking
Type "bo unpack demo-project.zip"
Sleep 600ms
Enter
Sleep 3000ms

# Show final result
Sleep 2000ms