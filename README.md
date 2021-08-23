# gcp-iap-tunnel-parser
Reverse engineering the IAP tunnel to WSS util


Huge shout out to https://github.com/GoogleCloudPlatform/iap-desktop for dealing with the en/decoding of messages.


The out of order message could be coming from the fact that I don't test -> wait -> connect local -> connect socket

```python
  def Run(self):
    """Start accepting connections."""
    if self._should_test_connection:
      try:
        self._TestConnection()
      except iap_tunnel_websocket.ConnectionCreationError as e:
        raise iap_tunnel_websocket.ConnectionCreationError(
            'While checking if a connection can be made: %s' % six.text_type(e))
    self._server_sockets = _OpenLocalTcpSockets(self._local_host,
                                                self._local_port)
    log.out.Print('Listening on port [%d].' % self._local_port)

    try:
      with execution_utils.RaisesKeyboardInterrupt():
        while True:
          self._connections.append(self._AcceptNewConnection())
          # To fix b/189195317, we will need to erase the reference of dead
          # tasks.
          self._CleanDeadClientConnections()
    except KeyboardInterrupt:
      log.info('Keyboard interrupt received.')
    finally:
      self._CloseServerSockets()

    self._shutdown = True
    self._CloseClientConnections()
    log.status.Print('Server shutdown complete.')
```