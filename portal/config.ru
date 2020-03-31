require_relative './app'

run Rack::URLMap.new({
                         "/" => Dashboard,
                         "/admin" => AdminConsole
                     })
