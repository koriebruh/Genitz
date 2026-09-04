# VHS tape for docs/demo.gif — regenerate with:
#   go build -o /tmp/genitz .
#   export PATH="/tmp:$PATH"
#   rm -rf /tmp/genitz-demo/shoply-api && mkdir -p /tmp/genitz-demo
#   cd /tmp/genitz-demo && vhs /path/to/docs/demo.tape
#
# Nothing staged — every command actually runs (real go get, real go.mod).
# Walks: genitz init -> folder/module -> dependency picker with a preset
# applied -> Docker toggle -> Extras (CI/README/Community files/license) ->
# Review -> real install -> a few of the newer standalone commands.

Output demo.gif

Require genitz

Set Shell "zsh"
Set FontSize 14
Set Width 1200
Set Height 900
Set Theme "Dracula"
Set Padding 20
Set TypingSpeed 60ms
Set WindowBar Colorful

Hide
Type "cd /tmp/genitz-demo && clear"
Enter
Sleep 300ms
Show

Type "genitz init"
Enter
Sleep 1.2s

Type "shoply-api"
Sleep 400ms
Enter
Sleep 600ms

Type "github.com/koriebruh/shoply-api"
Sleep 400ms
Enter
Sleep 800ms

# Dependency picker: apply the "Web API" preset (fiber + gorm + viper + zap)
Type "p"
Sleep 800ms
Enter
Sleep 1s
Enter
Sleep 1s

# Docker step
Type "y"
Sleep 500ms
Enter
Sleep 1s

# Extras: CI, README, Community files, license = mit
Space
Sleep 300ms
Down
Down
Down
Space
Sleep 300ms
Down
Space
Sleep 300ms
Down
Space
Sleep 800ms
Enter
Sleep 1.5s

# Review -> confirm -> real install
Enter
Sleep 20s

Hide
Type "clear"
Enter
Show
Sleep 300ms

Type "genitz search redis"
Enter
Sleep 2s

Type "genitz doctor"
Enter
Sleep 2.5s

Type "genitz preset list"
Enter
Sleep 3s
