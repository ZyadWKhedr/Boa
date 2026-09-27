# VHS Tape Script for Boa Terminal Demo
# Run: vhs docs/demo.tape

Output docs/img/demo.gif
Output docs/img/demo.mp4

Set FontSize 15
Set Width 1200
Set Height 760
Set Padding 20
Set Theme "Catppuccin Mocha"
Set WindowBar Colorful
Set TypingSpeed 35ms

# Launch interactive dashboard
Type "cd ~/Desktop && bo"
Sleep 600ms
Enter
Sleep 1800ms

# ==========================================
# 1. PACK: Compress ~/Desktop/boa-test
# ==========================================
Enter
Sleep 800ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 600ms
Enter
Sleep 1000ms
Enter
Sleep 2500ms
Enter
Sleep 1200ms

# ==========================================
# 2. LIST: Explore archive file contents
# ==========================================
Down
Sleep 250ms
Down
Sleep 600ms
Enter
Sleep 800ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 600ms
Enter
Sleep 3000ms
Enter
Sleep 1200ms

# ==========================================
# 3. INSPECT: Deep metadata analysis
# ==========================================
Down
Sleep 600ms
Enter
Sleep 800ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 600ms
Enter
Sleep 3000ms
Enter
Sleep 1200ms

# ==========================================
# 4. COMPARE: Multi-level visual comparison
# ==========================================
Down
Sleep 250ms
Down
Sleep 600ms
Enter
Sleep 800ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 600ms
Enter
Sleep 5000ms
Enter
Sleep 1200ms

# ==========================================
# 5. UNPACK: Safely extract archive
# ==========================================
Up
Sleep 250ms
Up
Sleep 250ms
Up
Sleep 250ms
Up
Sleep 600ms
Enter
Sleep 800ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 600ms
Enter
Sleep 1000ms
Enter
Sleep 2500ms
Enter
Sleep 1200ms

# ==========================================
# 6. STATUS: Runtime health & system info
# ==========================================
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 600ms
Enter
Sleep 2500ms
Enter
Sleep 1200ms

# ==========================================
# 7. QUIT: Graceful exit
# ==========================================
Type "q"
Sleep 1800ms