# frozen_string_literal: true

require 'open3'
require 'securerandom'
require 'time'

# Utility module for WebApp
module Utils
  def self.password_hash(salt, password)
    string = password + salt
    1000.times do
      string = `echo -n #{string} | openssl sha256`.chomp.split(' ')[1]
    end

    string
  end

  def self.password_salt
    SecureRandom.alphanumeric(128)
  end

  def self.today
    Time.now.strftime('%Y-%m-%d %H:%M:%S')
  end

  def self.session_string
    SecureRandom.alphanumeric(255)
  end
end
