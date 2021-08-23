#!/usr/bin/env ruby

require 'mqtt'
require 'date'
require 'securerandom'
require 'json'

BROKER = 'centos8-katello-4-0.oezr-fedora.example.com'
BROKER_PORT = 1883

CLIENT_UUID = '6deda2c6-244a-482c-84ea-9f940e88432c'

message = {
  type: 'data',
  message_id: SecureRandom.uuid,
  # client_uuid doesn't seemt to be used
  # client_uuid: CLIENT_UUID,
  version: 1,
  sent: DateTime.now.iso8601,
  directive: 'foreman',
  content: 'https://raw.githubusercontent.com/ezr-ondrej/foreman_ygg_worker/main/work',
  metadata: {
    return_url: 'http://raw.example.com/return'
  }
}

topic = "yggdrasil/#{CLIENT_UUID}/data/in"

MQTT::Client.connect(BROKER, BROKER_PORT) do |c|
  c.publish(topic, JSON.dump(message), false, 1)
end
