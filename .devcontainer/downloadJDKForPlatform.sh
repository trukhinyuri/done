#!/bin/bash

ARCH=$(uname -m)
OS=$(uname -s)

if [ "$OS" = "Linux" ]; then
  if [ "$ARCH" = "x86_64" ]; then
    JDK_URL="wget https://download.oracle.com/java/"$JDK_VERSION"/latest/jdk-"$JDK_VERSION"_linux-x64_bin.tar.gz"
  elif [ "$ARCH" = "aarch64" ]; then
      JDK_URL="wget https://download.oracle.com/java/"$JDK_VERSION"/latest/jdk-"$JDK_VERSION"_linux-aarch64_bin.tar.gz"
  else
    echo "Unsupported arch: $ARCH"
    exit 1
  fi
else
  echo "Unsupported OS: $OS"
  exit 1
fi

TMP_DIR=$(mktemp -d)
cd $TMP_DIR

echo "Download $JDK_URL"
curl -LO $JDK_URL
JDK_TAR=$(basename $JDK_URL)
tar -xzf $JDK_TAR

# Installing JDK
echo "Installing JDK..."
sudo mv jdk-$JDK_VERSION"* /usr/local/java
echo "export JAVA_HOME=/usr/local/java" >> ~/.bashrc
echo "export PATH=\$JAVA_HOME/bin:\$PATH" >> ~/.bashrc

# Очистка
rm -rf $TMP_DIR

echo "Установка Oracle JDK завершена."
