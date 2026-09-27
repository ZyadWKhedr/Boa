# VHS Tape Script for Boa Terminal Demo
# Install VHS: brew install vhs
# Run: vhs docs/demo.tape

Output docs/img/demo.gif
Output docs/img/demo.mp4

Set FontSize 15
Set Width 1200
Set Height 760
Set Padding 20
Set Theme "Catppuccin Mocha"
Set WindowBar Colorful
Set TypingSpeed 30ms

# 1. Interactive Dashboard
Type "bo"
Sleep 600ms
Enter
Sleep 1800ms
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
Up
Sleep 400ms
Type "q"
Sleep 800ms

Type "clear"
Enter
Sleep 400ms

# 2. Pack Desktop Folder
Type "bo pack ~/Desktop/boa-test -l 6"
Sleep 600ms
Enter
Sleep 2500ms

Type "clear"
Enter
Sleep 400ms

# 3. Inspect Archive Contents
Type "bo list ~/Desktop/boa-test.zip"
Sleep 600ms
Enter
Sleep 3500ms

Type "clear"
Enter
Sleep 400ms

# 4. Benchmark Deflate Levels (0-9) & Trade-offs
Type "bo bench ~/Desktop/boa-test --compare"
Sleep 600ms
Enter
Sleep 5000ms

Type "clear"
Enter
Sleep 400ms

# 5. Safely Unpack Archive
Type "bo unpack ~/Desktop/boa-test.zip -o ~/Desktop/boa-test-unpacked --force"
Sleep 600ms
Enter
Sleep 2500ms

Type "clear"
Enter
Sleep 400ms

# 6. System Status & Runtime Info
Type "bo status"
Sleep 600ms
Enter
Sleep 3000ms