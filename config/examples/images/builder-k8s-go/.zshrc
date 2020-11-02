export ZSH=$HOME/.oh-my-zsh

ZSH_THEME="robbyrussell"

plugins=(
  git
  kubectl
  docker
  kube-ps1
)

source $ZSH/oh-my-zsh.sh

source <(stern --completion=zsh)
source <(k completion zsh)
export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"

PROMPT=$PROMPT'$(kube_ps1) '