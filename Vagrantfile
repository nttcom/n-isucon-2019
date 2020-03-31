# -*- mode: ruby -*-
# vi: set ft=ruby :

APP_COUNT = 1

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/bionic64"

  (1..APP_COUNT).each do |node_id|
    config.vm.define "app#{node_id}" do |node|
      node.vm.hostname = "app#{node_id}"
      node.vm.network "private_network", ip: "192.168.33.#{10+node_id}"
      node.vm.network "forwarded_port", guest: 80, host: "#{8000+node_id}"
    end
  end

  config.vm.define :bench1 do |node|
    node.vm.box = config.vm.box
    node.vm.network "private_network", ip: "192.168.33.101"
  end

  config.vm.provider "virtualbox" do |vb|
    vb.gui = false
    vb.cpus = 2
    vb.memory = "2048"
  end

  config.vm.provision "ansible_local" do |ansible|
    ansible.playbook = "./infra/production/site.yml"
    ansible.install = true

    ansible.groups = {
      "app" => ["app[1:3]"],
      "app:vars" => {
                    "local_env" => true,
                    "local_app_dir" => "/vagrant/app"
      },
      "bench" => ["bench1"],
      "bench:vars" => {
                    "local_env" => true,
                    "local_scenario_dir" => "/vagrant/bench/scenario"
      },
      "all:vars" => {
                    "player_name" => "isucon",
                    "admin_name" => "isucon_admin",
                    "group_name" => "isucons",
                    "ansible_python_interpreter" => "/usr/bin/python3",
      }
    }
  end
end
