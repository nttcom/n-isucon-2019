root_dir = "#{Dir.getwd}"

bind "unix://#{root_dir}/tmp/puma.sock"
pidfile "#{root_dir}/tmp/puma.pid"
state_path "#{root_dir}/tmp/puma.state"
rackup "#{root_dir}/config.ru"

threads 4,8
activate_control_app
