#!/bin/sh
# 这是一个仓库代码同步工具

mkdir ../tmp_dir
cp -rf * ../tmp_dir
cd ../tmp_dir
rm -rf .git
mv README.md README_bak.md
cat README_bak.md | awk '{gsub(/github.com\/jmdrws\/gateway_demo\"); print $0}' > README_bak2.md
cat README_bak2.md
cat README_bak2.md | awk '{gsub(/github.com\/jmdrws/); print $0}' > README.md
cat README.md
rm README_bak2.md
rm README_bak.md
cp -rf * /Users/niuyufu/imooc/go_gateway
cd ../
rm -rf tmp_dir