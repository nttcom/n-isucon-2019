# frozen_string_literal: true

Bundler.require(:default, ENV['RACK_ENV'] || 'development')

require 'sinatra/reloader'
require 'rack/contrib'
require 'base64'

require_relative 'utils/utils.rb'

$db_params = {
  host: ENV.fetch('MYSQL_HOST'),
  username: ENV.fetch('MYSQL_USER'),
  password: ENV.fetch('MYSQL_PASSWORD'),
  database: ENV.fetch('MYSQL_DATABASE'),
  reconnect: true
}

Mysql2::Client.default_query_options.merge!(symbolize_keys: true)

ITEM_LIMIT = 10

# N-ISUCON WebApp class for Ruby
class App < Sinatra::Base
  use Rack::PostBodyContentTypeParser

  enable :sessions
  set :session_secret, 'for_sinatra_sessioneHFeqlgD4GEMqEsbpZFEY1ioFFHrOB'
  set :public_folder, File.dirname(__FILE__) + '/public'

  helpers do
    def logged_in?
      !session[:logged_in].nil?
    end

    def find(table:, id:)
      sql = "SELECT * FROM #{table} WHERE id=#{id}"
      client = Mysql2::Client.new($db_params)
      result = client.query(sql).to_a.first
      client.close

      result
    end

    def find_by_username(username)
      sql = "SELECT * FROM users WHERE username='#{username}'"
      client = Mysql2::Client.new($db_params)
      result = client.query(sql).to_a.first
      client.close

      result
    end

    def find_username(user_id)
      find(table: 'users', id: user_id)[:username]
    end
  end

  get '/' do
    redirect '/index.html'
  end

  post '/signin' do
    return 400 if params[:username].nil? || params[:password].nil?
    return 400 if params[:username].empty? || params[:password].empty?

    user = find_by_username(params[:username])
    return 401 unless user # user not found

    password_hash = Utils.password_hash(user[:salt], params[:password])
    return 401 unless password_hash == user[:password_hash] # wrong password

    session[:logged_in] = user[:id]

    json(username: user[:username])
  end

  get '/signout' do
    return 401 unless logged_in?

    session[:logged_in] = nil
    return 204
  end

  get '/users' do
    # debug endpoint
    results = Mysql2::Client.new($db_params).query('SELECT * from users').to_a
    p results.to_a

    return 200
  end

  post '/users' do
    p params
    return 400 if params[:username].nil? || params[:password].nil?

    user = find_by_username(params[:username])
    return 409 if user # existing user

    salt = Utils.password_salt
    password_hash = Utils.password_hash(salt, params[:password])
    current_time = Utils.today

    sql = 'INSERT INTO users ' \
          '(username, password_hash, salt, created_at, updated_at) ' \
          "VALUES ('#{params[:username]}', '#{password_hash}', '#{salt}', '#{current_time}', '#{current_time}')"
    client = Mysql2::Client.new($db_params)
    client.query(sql)

    response_body = {
      id: client.last_id,
      username: params[:username],
      created_at: current_time,
      updated_at: current_time
    }

    return 201, json(response_body)
  end

  get '/users/:name' do
    sql = 'SELECT id, username, created_at, updated_at FROM users '\
          "WHERE username='#{params[:name]}'"
    result = Mysql2::Client.new($db_params).query(sql).to_a.first

    return 404 if result.nil?

    json(result)
  end

  patch '/users/:name' do
    return 401 unless logged_in?

    result = find_by_username(params[:name])
    return 404 if result.nil?
    return 403 if result[:id] != session[:logged_in]

    username = params[:username]
    password = params[:password]
    return 400 if username.to_s.empty? && password.to_s.empty?

    if username
      return 409 if find_by_username(username)
    else
      username = result[:username]
    end

    if password.to_s.empty?
      salt = result[:salt]
      password_hash = result[:password_hash]
    else
      salt = Utils.password_salt
      password_hash = Utils.password_hash(salt, password)
    end

    sql = 'UPDATE users SET ' \
          "username='#{username}', password_hash='#{password_hash}', salt='#{salt}', " \
          "updated_at='#{Utils.today}' WHERE id=#{session[:logged_in]}"
    Mysql2::Client.new($db_params).query(sql)

    sql = 'SELECT id, username, created_at, updated_at FROM users '\
          "WHERE username='#{username}'"
    result = Mysql2::Client.new($db_params).query(sql).to_a.first

    return 200, json(result)
  end

  delete '/users/:name' do
    return 401 unless logged_in?

    result = find_by_username(params[:name])
    return 404 if result.nil?
    return 403 if result[:id] != session[:logged_in]

    sql = "DELETE FROM users WHERE id=#{result[:id]}"
    Mysql2::Client.new($db_params).query(sql)

    return 204
  end

  post '/users/:name/icon' do
    return 401 unless logged_in?

    user = find_by_username(params[:name])
    return 404 if user.nil?
    return 403 if user[:id] != session[:logged_in]

    icon = params[:iconimage]
    p params
    return 400 if icon.nil?

    sql = "SELECT * FROM icon WHERE user_id=#{user[:id]}"
    result = Mysql2::Client.new($db_params).query(sql).to_a.first

    return 409 if result

    file = icon['tempfile'].read
    sql = "INSERT INTO icon (user_id, icon) VALUES (#{user[:id]}, '#{Base64.encode64(file)}')"

    Mysql2::Client.new($db_params).query(sql)

    return 201
  end

  get '/users/:name/icon' do
    user = find_by_username(params[:name])
    return 404 if user.nil?

    sql = "SELECT * FROM icon WHERE user_id=#{user[:id]}"
    icon = Mysql2::Client.new($db_params).query(sql).to_a.first

    if icon
      content_type 'image/png'
      return Base64.decode64(icon[:icon])
    else
      send_file 'public/img/default_user_icon.png'
    end
  end

  get '/items' do
    offset = params['page'] ? params['page'].to_i : 0
    sorting = params['sort'] || nil

    sql = 'SELECT * FROM items'
    total_items = Mysql2::Client.new($db_params).query(sql).to_a.size

    return json(count: 0, items: []) if total_items.zero?

    sql = 'SELECT id, user_id, title, likes, created_at FROM items'
    result = Mysql2::Client.new($db_params).query(sql).to_a

    result.sort_by! { |r| r[:created_at] }.reverse!

    if sorting == 'like'
      result.sort_by! do |r|
        r[:likes] ? r[:likes].split(',').length : 0
      end.reverse!
    end

    result.each do |r|
      r[:username] = find_username(r[:user_id])
      r.delete(:user_id)
      r.delete(:likes)
    end
    json(count: total_items, items: result[offset * ITEM_LIMIT, ITEM_LIMIT])
  end

  post '/items' do
    return 401 unless logged_in?

    return 400 if params[:title].nil? || params[:body].nil?
    return 400 if params[:title].empty? || params[:body].empty?

    current_time = Utils.today
    sql = 'INSERT INTO items ' \
          '(user_id, title, body, created_at, updated_at) ' \
          "VALUES (#{session[:logged_in]}, '#{params[:title]}', '#{params[:body]}', '#{current_time}', '#{current_time}')"
    client = Mysql2::Client.new($db_params)
    client.query(sql)

    result = find(table: 'items', id: client.last_id)
    result[:username] = find_username(result[:user_id])
    result.delete(:user_id)
    result[:likes] ||= ''

    return 201, json(result)
  end

  get '/items/:id' do
    result = find(table: 'items', id: params[:id])

    return 404 if result.nil?

    result[:username] = find_username(result[:user_id])
    result.delete(:user_id)
    result[:likes] ||= ''
    json(result)
  end

  patch '/items/:id' do
    return 401 unless logged_in?

    target = find(table: 'items', id: params[:id])
    return 404 if target.nil?
    return 403 if target[:user_id] != session[:logged_in]
    return 400 if params[:title].nil? && params[:body].nil?

    title = params[:title].nil? ? target[:title] : params[:title]
    body = params[:body].nil? ? target[:body] : params[:body]

    return 400 if title.empty? || body.empty?

    current_time = Utils.today

    sql = 'UPDATE items SET '\
          "title='#{title}', body='#{body}', updated_at='#{current_time}' WHERE id=#{params[:id]}"
    Mysql2::Client.new($db_params).query(sql)

    target = find(table: 'items', id: params[:id])
    target[:username] = find_username(target[:user_id])
    target.delete(:user_id)
    target[:likes] ||= ''

    json(target)
  end

  delete '/items/:id' do
    return 401 unless logged_in?

    target = find(table: 'items', id: params[:id])
    return 404 if target.nil?
    return 403 if target[:user_id] != session[:logged_in]

    sql = "DELETE FROM items WHERE id=#{params[:id]}"
    Mysql2::Client.new($db_params).query(sql)

    return 204
  end

  get '/items/:id/likes' do
    item = find(table: 'items', id: params[:id])
    return 404 if item.nil?

    like_list = item[:likes]&.split(',')

    response_body = {
      like_count: like_list.nil? ? 0 : like_list.length,
      likes: item[:likes] || ''
    }

    json(response_body)
  end

  post '/items/:id/likes' do
    return 401 unless logged_in?

    item = find(table: 'items', id: params[:id])

    return 404 if item.nil?

    likes = []
    likes.concat(item[:likes]&.split(',')&.map(&:to_i) || [])

    user_id = session[:logged_in]
    username = find_username(user_id)

    likes << username unless likes.include?(username)

    sql = "UPDATE items SET likes='#{likes.join(',')}' WHERE id=#{params[:id]}"
    Mysql2::Client.new($db_params).query(sql)

    return 204
  end

  delete '/items/:id/likes' do
    return 401 unless logged_in?

    item = find(table: 'items', id: params[:id])
    return 404 if item.nil?

    user_id = session[:logged_in]
    username = find_username(user_id)

    likes = []
    likes.concat(item[:likes]&.split(',') || [])

    return 404 unless likes.include?(username)

    likes.delete(username)

    sql = "UPDATE items SET likes='#{likes.join(',')}' WHERE id=#{params[:id]}"
    Mysql2::Client.new($db_params).query(sql)

    return 204
  end

  get '/items/:id/comments' do
    item = find(table: 'items', id: params[:id])
    return 404 if item.nil?

    response = { item_id: item[:id], comments: [] }

    comments = find(table: 'comments', id: item[:id])

    comments&.each do |k, v|
      next if k == :id || v.nil?

      comment = JSON.parse(v)
      comment[:comment_id] = k.to_s.delete_prefix('comment_').to_i
      response[:comments] << comment
    end

    json(response)
  end

  post '/items/:id/comments' do
    return 401 unless logged_in?

    user_id = session[:logged_in]
    username = find(table: 'users', id: user_id)[:username]

    return 400 if params[:comment].nil? || params[:comment].empty?

    item = find(table: 'items', id: params[:id])
    return 404 if item.nil?

    exist_comments = find(table: 'comments', id: params[:id])
    new_comment = { username: username, comment: params[:comment] }

    comment_id = if exist_comments.nil?
                   1
                 else
                   exist_comments.keys.sort.each do |key|
                     next if key == :id || !exist_comments[key].nil?

                     break key.to_s.delete_prefix('comment_').to_i
                   end
                 end

    if comment_id == 1
      sql = "INSERT INTO comments (id, comment_#{format('%03d', comment_id)}) "\
        "VALUES (#{params[:id]}, '#{new_comment.to_json}')"
      Mysql2::Client.new($db_params).query(sql)
    else
      sql = 'UPDATE comments SET comment_'\
        "#{format('%03d', comment_id)}='#{new_comment.to_json}' WHERE id=#{params[:id]}"
      Mysql2::Client.new($db_params).query(sql)
    end

    response_body = new_comment
    response_body[:item_id] = params[:id]
    response_body[:comment_id] = comment_id
    response_body[:username] = username

    return 201, json(response_body)
  end

  delete '/items/:item_id/comments/:comment_id' do
    return 401 unless logged_in?

    user_id = session[:logged_in]

    comments = find(table: 'comments', id: params[:item_id])
    return 404 if comments.nil?

    column = "comment_#{format('%03d', params[:comment_id])}".to_sym
    return 404 if comments[column].nil?

    comment = JSON.parse(comments[column], symbolize_names: true)
    username = find(table: 'users', id: user_id)[:username]
    return 403 if comment[:username] != username

    sql = "UPDATE comments SET #{column}=NULL where id=#{params[:item_id]}"
    Mysql2::Client.new($db_params).query(sql)

    return 204
  end

  get '/initialize' do
    system('../common/db/init.sh')
    return 200
  end
end
