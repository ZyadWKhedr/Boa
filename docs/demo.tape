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

Type "bo"
Sleep 500ms
Enter
Sleep 1500ms

# Navigate menu with arrow keys
Down
Sleep 400ms
Down
Sleep 400ms
Up
Sleep 400ms
Up
Sleep 600ms

# Press Enter to open File Explorer
Enter
Sleep 1200ms

# Navigate file picker
Down
Sleep 400ms
Down
Sleep 400ms

# Select folder to compress
Enter
Sleep 1000ms

# Confirm compression
Enter
Sleep 2s

# Return to menu
Enter
Sleep 800ms

# Run inspect
Type "bo list cmd.zip"
Sleep 500ms
Enter
Sleep 2500ms

# Run benchmark
Type "bo bench ./cmd"
Sleep 500ms
Enter
Sleep 3s
