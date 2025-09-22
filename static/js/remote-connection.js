function main() {
  const tunnel = new Guacamole.WebSocketTunnel('ws://127.0.0.1:3000/ws/connect')
  const client = new Guacamole.Client(tunnel)
  const display = document.getElementById('display')

  display.appendChild(client.getDisplay().getElement());


  client.connect("height=" + window.innerHeight + "&width=" + window.innerWidth);

  let mouse = new Guacamole.Mouse(client.getDisplay().getElement());

  mouse.onmousedown =
    mouse.onmouseup =
    mouse.onmousemove = function (mouseState) {
      client.sendMouseState(mouseState);
    };

  let keyboard = new Guacamole.Keyboard(document);

  keyboard.onkeydown = function (keysym) {
    client.sendKeyEvent(1, keysym);
  };

  keyboard.onkeyup = function (keysym) {
    client.sendKeyEvent(0, keysym);
  };

  window.addEventListener('load', updateClipboard(client), true)
  window.addEventListener('copy', updateClipboard(client))
  window.addEventListener('cut', updateClipboard(client))
  window.addEventListener('focus', e => {
    if (e.target === window) {
      updateClipboard(client)
    }
  }, true)

  client.onclipboard = onClipboard

};

function updateClipboard(client) {
  readLocalClipboard().then(data => {
    writeRemoteClipboard(client, data, "text/plain")
  })
}

async function readLocalClipboard() {
  try {
    const data = await navigator.clipboard.readText()
    return data
  } catch (error) {
    console.log(error)
  }
}

async function writeLocalClipboard(data) {
  try {
    await navigator.clipboard.writeText(data)
  } catch (error) {
    console.log(error)
  }
}

function writeRemoteClipboard(client, data, mimeType) {
  const stream = client.createClipboardStream(mimeType)

  writer = new Guacamole.StringWriter(stream)
  writer.sendText(data)
  writer.sendEnd()

}

function onClipboard(stream, mimeType) {
  let reader = new Guacamole.StringReader(stream)
  let data = ''

  if (mimeType == "text/plain") {
    reader.ontext = text => {
      data += text
    }

    reader.onend = () => {
      writeLocalClipboard(data)
    }
  }

}

main()
