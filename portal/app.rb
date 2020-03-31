Bundler.require

require 'csv'
require 'sinatra/reloader'
require 'redis'
require 'pp'

DB = Sequel.connect("mysql2://#{ENV.fetch('DB_USER')}:#{ENV.fetch('DB_PASS')}@127.0.0.1:3306/dashboard")
REDIS = Redis.new(host: "#{ENV.fetch('REDIS_HOST')}", port: 6379)

unless DB.table_exists?(:teams)
  DB.create_table(:teams) do
    primary_key :id
    String :name
    String :password
    String :team_name, text: true
    String :target_server_pub_ip
    String :target_server_pri_ip
    String :extra_server1_pub_ip
    String :extra_server1_pri_ip
    String :extra_server2_pub_ip
    String :extra_server2_pri_ip
    String :server_username
    String :server_password
  end
end

unless DB.table_exists?(:results)
  DB.create_table(:results) do
    primary_key :id
    Int :team_id
    Int :score
    Int :status
    String :command_result, text: true
    DateTime :finished_date
  end
end

class Team < Sequel::Model
  plugin :password, column: :password
end

class Results < Sequel::Model
end

class AdminConsole < Sinatra::Base
  use Rack::Auth::Basic, "Admin area" do |user, pass|
    user == 'ncomadmin' && pass == ENV.fetch("ADMIN_PASSWORD")
  end

  get '/' do
    @teams = Team.all
    erb :admin
  end

  post '/register' do
    CSV.parse(params["team_information"]).each do |t|
      Team.create(
          name: t[0],
          password: t[1],
          team_name: t[2],
          server_username: t[3],
          server_password: t[4],
          target_server_pub_ip: t[5],
          target_server_pri_ip: t[6],
          extra_server1_pub_ip: t[7],
          extra_server1_pri_ip: t[8],
          extra_server2_pub_ip: t[9],
          extra_server2_pri_ip: t[10]
      )
    end
    redirect to '/'
  end
end

class Dashboard < Sinatra::Base
  enable :sessions

  get '/' do
    if session[:team_id]
      @team = Team[id: session[:team_id]]

      latest_result = DB["SELECT score, status, command_result, finished_date FROM results WHERE team_id = #{@team.id} ORDER BY id DESC LIMIT 1"].first

      if latest_result
        @latest_score = latest_result[:score]
        @latest_status = latest_result[:status] == 0 ? 'SUCCESS' : 'FAIL'
        t = latest_result[:finished_date]
        @finished_at = "#{t.hour}:#{t.min}:#{t.sec}"
        @command_result = latest_result[:command_result]
        @best_score = DB["SELECT score FROM results WHERE team_id = #{@team.id} ORDER BY score DESC LIMIT 1"].first[:score]
      end
    end
    erb :index
  end

  get '/request_bench' do
    redirect to '/login' unless session[:team_id]

    team = Team[id: session[:team_id]]

    # すでにRedisキューにリクエストがあったらうけつけない
    all_requests = REDIS.lrange('benchqueue', 0, -1)
    all_team_ids = all_requests.map do |request|
      # Queueには、"チームID,プライベートIP"(e.g. "1,10.0.0.1")の形式で保存してある
      request.split(',')[0].to_i
    end
    if all_team_ids.include?(team[:id])
      status 409
      'duplicate entry'
    else
      REDIS.lpush('benchqueue', "#{team[:id]},#{team[:target_server_pri_ip]}")
      'ok'
    end
  end

  get '/login' do
    erb :login
  end

  post '/login' do
    team = Team[name: params[:team_id]]

    unless team && team.authenticate(params[:password])
      redirect to '/login'
    end

    session[:team_id] = team.id
    redirect to '/'
  end

  get '/logout' do
    session[:team_id] = nil
    redirect to '/'
  end
end
