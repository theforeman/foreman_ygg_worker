BROKER = 'localhost'
BROKER_PORT = 1883

CLIENT_UUID = '6deda2c6-244a-482c-84ea-9f940e88432c'

message = {
  type: 'work',
  client_uuid: CLIENT_UUID,
  version: 1,
  sent: Time.now.iso8601,
  payload: {
    handler: 'foreman',
  }
}


MQTT::Client.connect(BROKER, BROKER_PORT) do |c|
  c.publish("per-host/#{@host}", JSON.dump(message), false, 1)
end
