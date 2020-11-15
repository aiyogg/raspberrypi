#!/bin/bash

if [[ ! $(cat /etc/shells) == *zsh* ]]; then
  echo "Start install zsh"
  apt install zsh
fi

chsh -s $(which zsh)

echo "Start install on-my-zsh"
sh -c "$(curl -fsSL https://raw.githubusercontent.com/robbyrussell/oh-my-zsh/master/tools/install.sh)"

# install plugins
zshplugins="plugins=(git git-open zsh-autosuggestions zsh-syntax-highlighting)"
sed -i "/^#/! s/plugins=.*/${zshplugins}/g" ~/.zshrc

echo "Start clone plugins"
git clone https://github.com/zsh-users/zsh-syntax-highlighting.git ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-syntax-highlighting
git clone https://github.com/zsh-users/zsh-autosuggestions ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-autosuggestions
git clone https://github.com/paulirish/git-open.git ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/git-open
echo "Please run 'source ~/.zshrc', then enjoy coding"

exit 0
