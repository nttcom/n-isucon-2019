# N-ISUCON 2019 Ansible playbook for the competition day
## requirement
- ansible >= 2.8
- cowsay :)

## how to execute
### before do
- SQLのダンプデータを取得する
- ↑を `seed.sql` にrenameし, roles/deploy_app/files/に置く

### exec ansible
```
vim hosts
# edit host address
ssh-add # for cloning from git repo

# decrypt secret info
echo "${password} > vaultpassword.txt"

ansible-vault decrypt group_vars/app.yml
ansible-vault decrypt group_vars/portal.app

ansible-playbook -i hosts site.yml
```

## hosts
- all
  - app
    - niita app
  - portal
  - bench
