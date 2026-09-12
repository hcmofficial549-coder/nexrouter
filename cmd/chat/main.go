package main

import (
"encoding/json"
	"log"
"net/http"
"os"
"os/signal"
"strings"
"time"

"github.com/gorilla/websocket"
	"github.com/nexrouter/nexrouter/auth"
"github.com/nexrouter/nexrouter/chat"
"github.com/nexrouter/nexrouter/core"
"github.com/nexrouter/nexrouter/middleware"
)

var (
hub = chat.NewHub()
	jwtSecret = initSecret()
upgrader = websocket.Upgrader{
ReadBufferSize:  1024,
WriteBufferSize: 1024,
CheckOrigin:     func(r *http.Request) bool { return true },
}
)

const chatUI = `<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>nexchat - realtime chat</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{background:#0a0a0f;color:#e4e4e7;font-family:-apple-system,'Segoe UI',sans-serif;height:100vh;display:flex;flex-direction:column}
#joinScreen{flex:1;display:flex;align-items:center;justify-content:center;padding:20px}
.join-card{background:#12121a;border:1px solid #27272a;border-radius:16px;padding:2rem;width:100%;max-width:380px;text-align:center}
.join-card h1{font-size:1.6rem;margin-bottom:.3rem;background:linear-gradient(135deg,#a855f7,#06b6d4);-webkit-background-clip:text;-webkit-text-fill-color:transparent}
.join-card p{color:#71717a;font-size:.78rem;margin-bottom:1.4rem}
.join-card input{width:100%;background:#0a0a0f;border:1px solid #27272a;border-radius:8px;padding:.7rem .9rem;color:#e4e4e7;font-size:.9rem;outline:none;margin-bottom:.8rem;font-family:inherit}
.join-card input:focus{border-color:#a855f7}
.join-card button{width:100%;padding:.75rem;border:none;border-radius:8px;background:linear-gradient(135deg,#a855f7,#06b6d4);color:#fff;font-weight:800;font-size:.95rem;cursor:pointer;font-family:inherit}
.divider{font-size:.66rem;color:#71717a;margin:.9rem 0 .7rem;text-transform:uppercase;letter-spacing:.05em}
.join-card button.gold{background:linear-gradient(135deg,#fbbf24,#f59e0b);color:#000}
.vbadge{display:inline-block;margin-left:5px;padding:0 5px;border-radius:4px;background:rgba(16,185,129,.2);color:#10b981;font-size:.58rem;font-weight:800;vertical-align:middle}
.rooms-hint{font-size:.7rem;color:#71717a;margin-top:.8rem}
#chatScreen{flex:1;display:none;flex-direction:column;height:100vh}
.chat-head{display:flex;align-items:center;gap:.75rem;padding:.8rem 1.2rem;background:#12121a;border-bottom:1px solid #27272a}
.chat-head .room{font-weight:800;color:#06b6d4}
.chat-head .online{font-size:.72rem;color:#10b981}
.chat-head .spacer{flex:1}
.chat-head button{background:#1a1a2e;border:1px solid #27272a;color:#a1a1aa;padding:.35rem .8rem;border-radius:6px;cursor:pointer;font-size:.75rem;font-family:inherit}
#msgs{flex:1;overflow-y:auto;padding:1rem;display:flex;flex-direction:column;gap:.4rem}
.sys{align-self:center;font-size:.72rem;color:#71717a;font-style:italic;padding:.15rem .7rem}
.msg{max-width:75%;padding:.5rem .8rem;border-radius:12px;font-size:.85rem;display:flex;flex-direction:column;gap:.15rem}
.msg.them{align-self:flex-start;background:#1a1a2e;border:1px solid #27272a}
.msg.me{align-self:flex-end;background:linear-gradient(135deg,#7c3aed,#0891b2);color:#fff}
.msg .who{font-size:.66rem;font-weight:800;color:#a855f7}
.msg.me .who{color:#e9d5ff}
.msg .txt{word-break:break-word;white-space:pre-wrap}
.msg .t{font-size:.6rem;opacity:.55;align-self:flex-end}
.typing-bar{padding:.15rem 1rem;font-size:.72rem;color:#a855f7;font-style:italic;min-height:1.1rem}
.chat-input{display:flex;gap:.5rem;padding:.8rem 1rem;background:#12121a;border-top:1px solid #27272a}
.chat-input input{flex:1;background:#0a0a0f;border:1px solid #27272a;border-radius:8px;padding:.65rem .9rem;color:#e4e4e7;font-size:.88rem;outline:none;font-family:inherit}
.chat-input input:focus{border-color:#a855f7}
.chat-input button{padding:.65rem 1.2rem;border:none;border-radius:8px;background:linear-gradient(135deg,#a855f7,#06b6d4);color:#fff;font-weight:800;cursor:pointer;font-family:inherit}
</style>
</head>
<body>
<div id="joinScreen">
<div class="join-card">
<h1>nexchat</h1>
<p>realtime chat - powered by nexrouter + WebSocket</p>
<input id="nameInput" placeholder="Nama Anda" maxlength="24">
<input id="roomInput" placeholder="Room (default: general)" maxlength="24">
<button onclick="joinRoom()">Join as Guest
- atau -


Login and Join</button>
<div class="rooms-hint" id="roomsHint">loading rooms...</div>
</div>
</div>
<div id="chatScreen">
<div class="chat-head">
<span class="room" id="roomTitle">#general</span>
<span class="online" id="online">0 online</span>
<span class="spacer"></span>
<button onclick="leaveChat()">Exit</button>
</div>
<div id="msgs"></div>
<form class="chat-input" id="chatForm">
<input id="msgInput" autocomplete="off" placeholder="Ketik pesan..." maxlength="500">
<button type="submit">Send</button>
</form>
</div>
<script>
var ws = null, myName = '', myRoom = '', allowReconnect = true;
var authToken = '';
var API_URL = (location.hostname === 'localhost' || location.hostname === '127.0.0.1') ? 'http://localhost:8080' : 'https://nexrouter.up.railway.app';

function loadRoomsHint(){
  fetch('/rooms').then(function(r){return r.json();}).then(function(d){
    var list = (d.data||[]).map(function(x){return x.name+' ('+x.online+')';}).join(', ');
    document.getElementById('roomsHint').textContent = 'Rooms: ' + (list || 'general');
  }).catch(function(){});
}
loadRoomsHint();

function loginJoin(){
  var email = document.getElementById('loginEmail').value.trim();
  var pass = document.getElementById('loginPass').value;
  if(!email || !pass){ alert('Email dan password wajib diisi'); return; }
  myRoom = document.getElementById('roomInput').value.trim() || 'general';
  fetch(API_URL + '/api/v1/auth/login', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({email: email, password: pass})
  })
  .then(function(r){ return r.json().then(function(d){ return {ok: r.ok, d: d}; }); })
  .then(function(res){
    if(!res.ok){ alert('Login gagal: ' + ((res.d && res.d.error) || 'cek email/password')); return; }
    authToken = res.d.token;
    myName = (res.d.user && res.d.user.name) || email;
    allowReconnect = true;
    loadHistoryThen(connectWS);
  })
  .catch(function(e){ alert('API nexrouter tidak terjangkau di ' + API_URL + ' - pastikan API running. Detail: ' + e.message); });
}

function joinRoom(){
  authToken = '';
  myName = document.getElementById('nameInput').value.trim() || 'anonymous';
  myRoom = document.getElementById('roomInput').value.trim() || 'general';
  allowReconnect = true;
  loadHistoryThen(connectWS);
}

function loadHistoryThen(next){
  fetch('/rooms/'+encodeURIComponent(myRoom)+'/history')
    .then(function(r){return r.json();})
    .then(function(d){
      document.getElementById('msgs').innerHTML='';
      (d.data||[]).forEach(addMsg);
      next();
    })
    .catch(next);
}

function connectWS(){
  var proto = location.protocol === 'https:' ? 'wss://' : 'ws://';
  ws = new WebSocket(proto + location.host + '/ws?name=' + encodeURIComponent(myName) + '&room=' + encodeURIComponent(myRoom) + (authToken ? '&token=' + encodeURIComponent(authToken) : ''));
  ws.onopen = function(){
    document.getElementById('joinScreen').style.display='none';
    document.getElementById('chatScreen').style.display='flex';
    document.getElementById('roomTitle').textContent='#'+myRoom;
    updateOnline();
    setInterval(updateOnline, 5000);
  };
  ws.onmessage = function(e){
    try { var mm = JSON.parse(e.data); if(mm.type === 'typing'){ showTyping(mm.from); } else { hideTyping(); addMsg(mm); } } catch(err){}
  };
  ws.onclose = function(){
    if(allowReconnect){
      addMsg({type:'system', text:'Koneksi terputus - menyambung ulang...', time:''});
      setTimeout(connectWS, 2000);
    }
  };
}

function updateOnline(){
  fetch('/rooms').then(function(r){return r.json();}).then(function(d){
    var rooms=d.data||[];
    for(var i=0;i<rooms.length;i++){
      if(rooms[i].name===myRoom){
        document.getElementById('online').textContent=rooms[i].online+' online';
      }
    }
  }).catch(function(){});
}

var typingTimer = null, lastTypingSent = 0;
function showTyping(who){ var el = document.getElementById("typingInd"); el.textContent = who + " is typing..."; clearTimeout(typingTimer); typingTimer = setTimeout(function(){ el.textContent = ""; }, 2500); }
function hideTyping(){ document.getElementById("typingInd").textContent = ""; }
function sendTyping(){ var n = Date.now(); if(ws && ws.readyState === 1 && n - lastTypingSent > 1500){ lastTypingSent = n; ws.send(JSON.stringify({type:"typing"})); } }
function addMsg(m){
  var box=document.getElementById('msgs');
  var div=document.createElement('div');
  if(m.type==='message'){
    div.className='msg '+(m.from===myName?'me':'them');
    var who=document.createElement('span'); who.className='who'; who.textContent=m.from; if(m.verified){ var vb=document.createElement('span'); vb.className='vbadge'; vb.textContent='verified'; who.appendChild(vb); }
    var txt=document.createElement('span'); txt.className='txt'; txt.textContent=m.text;
    var t=document.createElement('span'); t.className='t'; t.textContent=m.time||'';
    div.appendChild(who); div.appendChild(txt); div.appendChild(t);
  } else {
    div.className='sys';
    div.textContent=(m.time?('['+m.time+'] '):'')+(m.text||'');
  }
  box.appendChild(div);
  box.scrollTop=box.scrollHeight;
  while(box.children.length>300){ box.removeChild(box.firstChild); }
}

document.getElementById('msgInput').addEventListener('input', sendTyping);
document.getElementById('chatForm').addEventListener('submit', function(e){
  e.preventDefault();
  var inp=document.getElementById('msgInput');
  var txt=inp.value.trim();
  if(!txt||!ws||ws.readyState!==1) return;
  ws.send(JSON.stringify({type:"chat", text: txt}));
  inp.value='';
  inp.focus();
});

function leaveChat(){
  allowReconnect=false;
  if(ws) ws.close();
  location.reload();
}
</script>
</body>
</html>
`
func main() {
dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/nexchat.db"
	}
	if err := chat.InitStore(dbPath); err != nil {
		log.Printf("[nexchat] WARNING: persistence off: %v", err)
	} else {
		log.Printf("[nexchat] persistence: %s", dbPath)
	}

	r := core.New()
r.Use(middleware.Recovery())
r.Use(middleware.CORS())

r.GET("/health", func(c *core.Context) {
c.JSON(http.StatusOK, core.H{"status": "ok", "app": "nexchat", "version": "1.2.0"})
})

r.GET("/", func(c *core.Context) {
c.SetHeader("Content-Type", "text/html; charset=utf-8")
c.SetHeader("Cache-Control", "no-cache")
c.Writer.WriteHeader(http.StatusOK)
c.Writer.Write([]byte(chatUI))
})

r.GET("/rooms", func(c *core.Context) {
c.JSON(http.StatusOK, core.H{"success": true, "data": hub.Rooms()})
})

r.GET("/rooms/:room/history", func(c *core.Context) {
room := c.Param("room")
c.JSON(http.StatusOK, core.H{"success": true, "room": room, "data": historyOrLoad(room)})
})

r.GET("/ws", wsHandler)

go func() {
quit := make(chan os.Signal, 1)
signal.Notify(quit, os.Interrupt)
<-quit
log.Println("[nexchat] bye")
os.Exit(0)
}()

port := os.Getenv("PORT")
if port == "" {
port = "8081"
}
log.Printf("[nexchat] v1.2.0 starting on :%s (powered by nexrouter!)", port)
if err := r.Run(":" + port); err != nil {
log.Fatal(err)
}
}

func wsHandler(c *core.Context) {
name := strings.TrimSpace(c.Query("name"))
room := strings.TrimSpace(c.Query("room"))
if name == "" {
name = "anonymous"
}
if len(name) > 24 {
name = name[:24]
}
if room == "" {
room = "general"
}

conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
if err != nil {
log.Printf("[nexchat] ws upgrade failed: %v", err)
return
}

send := make(chan []byte, 32)
verified := false
	if tok := strings.TrimSpace(c.Query("token")); tok != "" {
		if claims, err := auth.ValidateToken(tok, jwtSecret); err == nil {
			verified = true
			if claims.Name != "" {
				name = claims.Name
			}
			log.Printf("[nexchat] verified user: %s (%s)", name, claims.Email)
		} else {
			log.Printf("[nexchat] invalid token: %v", err)
		}
	}
	client := hub.Join(name, room, send)
	client.Verified = verified
log.Printf("[nexchat] %s joined #%s (online: %d)", name, room, hub.Online(room))

// write pump + keepalive ping (menjaga koneksi di balik proxy Railway)
go func() {
ticker := time.NewTicker(30 * time.Second)
defer ticker.Stop()
for {
select {
case msg, ok := <-send:
conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
if !ok {
conn.WriteMessage(websocket.CloseMessage, []byte{})
conn.Close()
return
}
if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
conn.Close()
return
}
case <-ticker.C:
conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
conn.Close()
return
}
}
}
}()

// read pump (blocking sampai koneksi putus)
conn.SetReadLimit(2048)
conn.SetReadDeadline(time.Now().Add(75 * time.Second))
conn.SetPongHandler(func(string) error {
conn.SetReadDeadline(time.Now().Add(75 * time.Second))
return nil
})
for {
_, raw, err := conn.ReadMessage()
if err != nil {
break
}
text := strings.TrimSpace(string(raw))
if text == "" {
continue
}
if len(text) > 2000 {
text = text[:500]
}
handleClientMessage(client, text)
}

hub.Leave(client)
close(send)
log.Printf("[nexchat] %s left #%s", name, room)
}
func handleClientMessage(c *chat.Client, raw string) {
var env struct {
Type string `json:"type"`
Text string `json:"text"`
}
if err := json.Unmarshal([]byte(raw), &env); err == nil {
if env.Type == "typing" {
hub.BroadcastTyping(c)
return
}
if env.Type == "chat" {
raw = env.Text
}
}
raw = strings.TrimSpace(raw)
if raw == "" {
return
}
if len(raw) > 500 {
raw = raw[:500]
}
hub.Broadcast(c, raw)
}

func historyOrLoad(room string) []chat.Message {
hist := hub.History(room)
if len(hist) > 0 {
return hist
}
return chat.LoadHistory(room, 50)
}
func initSecret() string {
if s := os.Getenv("JWT_SECRET"); s != "" {
return s
}
return "nexrouter-dev-secret-change-me"
}