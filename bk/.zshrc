export ZSH=$HOME/.oh-my-zsh

# Set name of the theme to load --- if set to "random", it will
# load a random theme each time oh-my-zsh is loaded, in which case,
# to know which specific one was loaded, run: echo $RANDOM_THEME
# See https://github.com/ohmyzsh/ohmyzsh/wiki/Themes
# ZSH_THEME="robbyrussell"
# ZSH_THEME_RANDOM_CANDIDATES=( "robbyrussell" "agnoster" "cloud" "nicoulaj" "nanotech" "juanghurtado" "base" "")
# ZSH_THEME="spaceship"
ZSH_THEME="nanotech"
# ZSH_THEME="cloud"
# ZSH_THEME="juanghurtado"
# ZSH_THEME="nicoulaj"
# ZSH_THEME="robbyrussell"
# ZSH_THEME="base"

# Set list of themes to pick from when loading at random
# Setting this variable when ZSH_THEME=random will cause zsh to load
# a theme from this variable instead of looking in $ZSH/themes/
# If set to an empty array, this variable will have no effect.
# ZSH_THEME_RANDOM_CANDIDATES=( "robbyrussell" "agnoster" )
# ZSH_THEME="ma"
# ZSH_THEME="cloud"
plugins=(
  git 
  zsh-autosuggestions 
  zsh-syntax-highlighting 
  bgnotify
  golang
  zsh-fzf-history-search
  zsh-autocomplete
)

PATH="$PREFIX/bin:$HOME/.local/bin:$PATH"
export PATH

LINK="https://github.com/mayTermux"
export LINK

LINK_SSH="git@github.com:mayTermux"
export LINK_SSH

export TERM=xterm-256color 

source $ZSH/oh-my-zsh.sh
source $HOME/.config/lf/icons

# Prediction List View
# zstyle ':autocomplete:*' default-context history-incremental-search-backward
# zstyle ':autocomplete:history-incremental-search-backward:*' min-input 1

source $HOME/func.zsh
set_editor
export EDITOR="nvim"
source $HOME/.aliases
source $HOME/.autostart

# pnpm
export PNPM_HOME="/data/data/com.termux/files/home/.local/share/pnpm"
case ":$PATH:" in
  *":$PNPM_HOME:"*) ;;
  *) export PATH="$PNPM_HOME:$PATH" ;;
esac
# pnpm end

fpath+=~/.zfunc; autoload -Uz compinit; compinit
export PATH=$PATH:/data/data/com.termux/files/home/.cargo/bin
autoload -U compinit; compinit
 export PATH=/data/data/com.termux/files/home/.local/share/pnpm:/data/data/com.termux/files/usr/bin:/data/data/com.termux/files/home/.local/bin:/data/data/com.termux/files/usr/bin:/data/data/com.termux/files/home/.cargo/bin:~/go/bin
export "GOPATH=$HOME/go"
export GOBIN="$GOPATH/bin"
alias interface_any="find . -type f -name '*' | xargs sed -i 's/interface{}/any/g'"

alias cat="/data/data/com.termux/files/usr/bin/cat"
source ~/func.zsh
export PATH=$PATH:$HOME/.local/bin/lvim 
export PYENV_ROOT="$HOME/.pyenv"
[[ -d $PYENV_ROOT/bin ]] && export PATH="$PYENV_ROOT/bin:$PATH"
eval "$(pyenv init - zsh)"

# opencode
export PATH=/data/data/com.termux/files/home/.opencode/bin:$PATH
export GOPRIVATE="github.com/niiniyare/erp,github.com/project/erp"
# export GEMINI_API_KEY="AIzaSyCIAtsw2MiZy-xZFX1UK96gzI4vPmNSpcs"
 export GEMINI_API_KEY="AQ.Ab8RN6LCCKAZd0cHtNSh-B2TjA67ej6Bkrm8SYh_oMUqKDmhyw"
# export GEMINI_API_KEY="AIzaSyDcoF9NJ0QPRDxRZQn0ky4COuxtrIpf2WQ"
export GEMINI_MODEL="gemini-2.5-pro"
# Auto-start Redis if not running
if ! redis-cli ping >/dev/null 2>&1; then
    redis-server --daemonize yes
fi
export GH_token='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'
export deepseek="sk-ab90583da8f44d44ae7d62afefcdd93c"
alias ubu="proot-distro login --user root --termux-home --shared-tmp -- zsh ubuntu"
alias fetch="neofetch"

# Auto-start Temporal if not running
if ! pgrep -f "temporal server" >/dev/null 2>&1; then
    temporal server start-dev >/dev/null 2>&1 &
fi
export ROLLUP_USE_NATIVE=false


export PATH=~/.npm-global/bin:$PATH

alias claude="~/.npm-global/bin/claude"
export testStripeKey="sk_test_4eC39HqLyjWDarjtT1zdp7dc"
export DATABASE_URL="postgres://admin:admin@localhost:5432/awo?sslmode=disable"
export REDIS_URL="redis://localhost:6379"
