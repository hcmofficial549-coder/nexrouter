package main

import (
"bytes"
	"encoding/json"
	"log"
"net/http"
"os"
"os/signal"
"strconv"
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
	spamGuard = chat.NewSpamGuard(10, 5*time.Second)
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
.msg.pm{border:1px dashed #ec4899}
.msg.pm .who{color:#ec4899}
.vbadge.adm{background:rgba(251,191,36,.18);color:#fbbf24}
.cmd-hint{padding:.3rem 1rem;font-size:.64rem;color:#71717a;background:#0d1117;border-top:1px solid #27272a;text-align:center}
.rx-picker{position:absolute;top:-36px;right:4px;display:flex;gap:2px;background:#1a1a2e;border:1px solid #3f3f46;border-radius:9px;padding:3px;z-index:50;box-shadow:0 6px 16px rgba(0,0,0,.55)}
.rx-picker button{background:none;border:none;font-size:1.05rem;cursor:pointer;padding:2px 4px;border-radius:6px}
.rx-picker button:hover{background:#27272a;transform:scale(1.25)}
.rx-bar{display:flex;gap:4px;flex-wrap:wrap;margin-top:4px}
.rx-chip{background:rgba(168,85,247,.15);border:1px solid rgba(168,85,247,.35);border-radius:999px;padding:0 8px;font-size:.68rem}
.msg{cursor:pointer}
.sidebar{position:fixed;left:0;top:0;bottom:0;width:190px;background:#12121a;border-right:1px solid #27272a;padding:.8rem .6rem;overflow-y:auto;z-index:40;display:none}
.sidebar.show{display:block}
body.sb-on{padding-left:200px}
.sb-logo{font-weight:900;color:#06b6d4;font-size:.95rem;margin-bottom:.6rem;padding:0 .3rem}
.sb-label{font-size:.6rem;color:#71717a;text-transform:uppercase;letter-spacing:.08em;margin:.6rem .3rem .3rem}
.room-item{display:flex;align-items:center;gap:.4rem;padding:.45rem .55rem;border-radius:7px;cursor:pointer;font-size:.8rem;color:#a1a1aa;margin-bottom:.15rem}
.room-item:hover{background:#1a1a2e;color:#e4e4e7}
.room-item.active{background:linear-gradient(135deg,#7c3aed,#0e7490);color:#fff;font-weight:700}
.room-item .rn{flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.room-item .ro{font-size:.65rem;color:#10b981}
.room-item .rb{background:#ef4444;color:#fff;font-size:.6rem;font-weight:800;border-radius:999px;min-width:17px;height:17px;display:inline-flex;align-items:center;justify-content:center;padding:0 4px}
.dm-item{display:flex;align-items:center;gap:.4rem;padding:.4rem .55rem;border-radius:7px;cursor:pointer;font-size:.78rem;color:#a1a1aa;margin-bottom:.15rem}
.dm-item:hover{background:#1a1a2e;color:#e4e4e7}
.dm-item.active{background:linear-gradient(135deg,#be185d,#9333ea);color:#fff;font-weight:700}
.dm-item .dn{flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.dm-item .dt{font-size:.6rem;color:#71717a}
.dt-empty{font-size:.65rem;color:#71717a;padding:.2rem .55rem;font-style:italic}
.sb-newroom{width:100%;margin-top:.4rem;padding:.45rem;background:#1a1a2e;border:1px dashed #3f3f46;border-radius:7px;color:#a1a1aa;font-size:.75rem;cursor:pointer;font-family:inherit}
.sb-newroom:hover{border-color:#06b6d4;color:#06b6d4}
.sb-foot{margin-top:1rem;padding:.5rem .3rem;border-top:1px solid #27272a;font-size:.7rem;color:#71717a}
@media(max-width:640px){body.sb-on{padding-left:0}.sidebar{display:none}}
.btn-del{background:none;border:none;color:#ef4444;cursor:pointer;font-size:.7rem;padding:0 4px;opacity:0.6;margin-left:4px}
.btn-del:hover{opacity:1}
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
<div class="sidebar" id="sidebar">
<div class="sb-logo">nexchat</div>
<div class="sb-label">Rooms</div>
<div id="roomList"></div>
<button class="sb-newroom" onclick="newRoom()">+ new room</button>
<div class="sb-label">Direct</div>
<div id="dmList"></div>
<div class="sb-foot" id="sbUser"></div>
</div>
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
<div class="cmd-hint">commands: /pm "username" pesan | /history "username" | admin: /kick "username"</div>
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
    document.getElementById('chatScreen').style.display='flex'; document.getElementById('sidebar').classList.add('show'); document.body.classList.add('sb-on'); document.getElementById('sbUser').textContent = myName;
    document.getElementById('roomTitle').textContent='#'+myRoom;
    updateRooms(); sendDMList();
    if(!window.__roomsTimer){ window.__roomsTimer = setInterval(updateRooms, 5000); }
  };
  ws.onmessage = function(e){
    try { var mm = JSON.parse(e.data); if(mm.type === 'typing'){ showTyping(mm.from); } else if(mm.type === 'kick'){ allowReconnect = false; alert('Anda di-kick: ' + (mm.reason || '')); location.reload(); } else if(mm.type === 'pm'){ hideTyping(); addPM(mm); } else if(mm.type === 'react'){ renderReactions(mm.msg_id, mm.reactions); } else if(mm.type === 'dm_list'){ renderDMs(mm.threads); } else if(mm.type === 'dm_history'){ renderDMHistory(mm); } else { hideTyping(); if(!dmTarget || mm.from === 'system'){ addMsg(mm); } } } catch(err){}
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
function addPM(m){
  var box=document.getElementById('msgs');
  var div=document.createElement('div');
  var mine = (m.from === myName);
  div.className='msg pm ' + (mine?'me':'them');
  var who=document.createElement('span'); who.className='who';
  who.textContent = mine ? ('PM to @' + m.to) : ('PM from @' + m.from);
  var txt=document.createElement('span'); txt.className='txt'; txt.textContent=m.text;
  var t=document.createElement('span'); t.className='t'; t.textContent=m.time||'';
  div.appendChild(who); div.appendChild(txt); div.appendChild(t); if(m.id){ div.dataset.msgid = m.id; div.style.position='relative'; div.onclick = function(e){ if(!e.target.classList.contains('btn-del')){ showPicker(m.id, div); } }; if(m.from === myName){ var db = document.createElement('button'); db.className='btn-del'; db.innerHTML='\uD83D\uDDD1\uFE0F'; db.title='Delete'; db.onclick = (function(mid){ return function(ev){ ev.stopPropagation(); if(confirm('Hapus?')){ ws.send(JSON.stringify({type:'chat', text:'/delete '+mid})); } }; })(m.id); who.appendChild(db); } }
  box.appendChild(div);
  box.scrollTop=box.scrollHeight;
}
var rxPicker = null;
var RX_EMOJIS = ['\uD83D\uDC4D', '\u2764\uFE0F', '\uD83D\uDE02', '\uD83C\uDF89', '\uD83D\uDD25'];
function showPicker(msgId, bubble){
  hidePicker();
  rxPicker = document.createElement('div');
  rxPicker.className = 'rx-picker';
  RX_EMOJIS.forEach(function(em){
    var b = document.createElement('button');
    b.textContent = em;
    b.onclick = function(e){
      e.stopPropagation();
      if(ws && ws.readyState === 1){
        ws.send(JSON.stringify({type:'react', msg_id: msgId, emoji: em}));
      }
      hidePicker();
    };
    rxPicker.appendChild(b);
  });
  bubble.appendChild(rxPicker);
  setTimeout(function(){ document.addEventListener('click', hidePickerOnce); }, 20);
}
function hidePickerOnce(){ hidePicker(); document.removeEventListener('click', hidePickerOnce); }
function hidePicker(){
  if(rxPicker && rxPicker.parentNode){ rxPicker.parentNode.removeChild(rxPicker); }
  rxPicker = null;
}
function renderDMHistory(mm){
  var msgs = mm.messages || [];
  addMsg({type:'system', text:'--- DM history with @' + mm['with'] + ' (' + msgs.length + ' pesan tersimpan) ---', time:''});
  for(var i=0; i<msgs.length; i++){ addPM(msgs[i]); }
}
function renderReactions(msgId, reactions){
  var bubbles = document.querySelectorAll('[data-msgid="' + msgId + '"]');
  for(var i=0; i<bubbles.length; i++){
    var b = bubbles[i];
    var old = b.querySelector('.rx-bar');
    if(old){ old.parentNode.removeChild(old); }
    var keys = Object.keys(reactions || {});
    if(!keys.length){ continue; }
    var bar = document.createElement('div');
    bar.className = 'rx-bar';
    keys.forEach(function(em){
      var chip = document.createElement('span');
      chip.className = 'rx-chip';
      chip.textContent = em + ' ' + reactions[em].length;
      chip.title = reactions[em].join(', ');
      bar.appendChild(chip);
    });
    b.appendChild(bar);
  }
}
var unread = {};
var lastActSeen = {};
function renderRooms(rooms){
  var el = document.getElementById('roomList');
  el.innerHTML = '';
  rooms.forEach(function(r){
    var d = document.createElement('div');
    d.className = 'room-item' + (r.name === myRoom ? ' active' : '');
    var nm = document.createElement('span'); nm.className='rn'; nm.textContent = '#' + r.name;
    d.appendChild(nm);
    if(unread[r.name] && r.name !== myRoom){
      var b = document.createElement('span'); b.className='rb';
      b.textContent = unread[r.name] > 99 ? '99+' : unread[r.name];
      d.appendChild(b);
    } else {
      var on = document.createElement('span'); on.className='ro'; on.textContent = r.online;
      d.appendChild(on);
    }
    d.onclick = (function(name){ return function(){ switchRoom(name); }; })(r.name);
    el.appendChild(d);
  });
}
function newRoom(){
  var name = prompt('New room name:');
  if(!name) return;
  name = name.trim().toLowerCase().replace(/[^a-z0-9\-_]/g, '-').substring(0, 24);
  if(!name) return;
  switchRoom(name);
}
var dmTarget = null;
var myAdmin = false;
function sendDMList(){ if(ws && ws.readyState === 1){ ws.send(JSON.stringify({type:'dm_list'})); } }
function renderDMs(threads){
  var el = document.getElementById('dmList');
  if(!el) return;
  el.innerHTML = '';
  if(!threads || !threads.length){
    var e = document.createElement('div');
    e.className = 'dt-empty';
    e.textContent = 'belum ada DM - kirim /pm dulu';
    el.appendChild(e);
    return;
  }
  threads.forEach(function(t){
    var d = document.createElement('div');
    d.className = 'dm-item' + (dmTarget === t['with'] ? ' active' : '');
    var nm = document.createElement('span'); nm.className='dn'; nm.textContent = '@' + t['with'];
    var tm = document.createElement('span'); tm.className='dt'; tm.textContent = t.last || '';
    d.appendChild(nm); d.appendChild(tm);
    d.onclick = (function(name){ return function(){ openDM(name); }; })(t['with']);
    el.appendChild(d);
  });
}
function openDM(name){
  dmTarget = name;
  document.getElementById('roomTitle').textContent = '@' + name + ' (DM)';
  document.getElementById('msgs').innerHTML = '';
  if(ws && ws.readyState === 1){
    sendDMList();
    ws.send('/history "' + name + '"');
  }
}
function exitDM(){
  dmTarget = null;
  document.getElementById('roomTitle').textContent = '#' + myRoom;
  document.getElementById('msgs').innerHTML = '';
  loadHistoryThen(function(){ sendDMList(); });
}
function switchRoom(name){
  if(name === myRoom){ if(dmTarget){ exitDM(); } return; }
  dmTarget = null;
  unread[name] = 0;
  myRoom = name;
  allowReconnect = false;
  if(ws){ ws.close(); }
  document.getElementById('msgs').innerHTML = '';
  document.getElementById('roomTitle').textContent = '#' + myRoom;
  loadHistoryThen(function(){ allowReconnect = true; connectWS(); });
}
function updateRooms(){
  fetch('/rooms').then(function(r){return r.json();}).then(function(d){
    var rooms = d.data || [];
    rooms.forEach(function(r){
      var prev = lastActSeen[r.name] || 0;
      if(r.last_act && r.last_act > prev){
        if(prev !== 0 && r.name !== myRoom){ unread[r.name] = (unread[r.name]||0) + 1; }
        lastActSeen[r.name] = r.last_act;
      }
      if(r.name === myRoom){
        unread[r.name] = 0;
        document.getElementById('online').textContent = r.online + ' online';
      }
    });
    renderRooms(rooms);
  }).catch(function(){});
}
function removeMsg(id){
  var els = document.querySelectorAll('[data-msgid="' + id + '"]');
  for(var i=0; i<els.length; i++){
    els[i].style.opacity = '0.3';
    els[i].style.textDecoration = 'line-through';
    setTimeout((function(el){ return function(){ el.remove(); }; })(els[i]), 500);
  }
}

function addMsg(m){
  var box=document.getElementById('msgs');
  var div=document.createElement('div');
  if(m.type==='message'){
    div.className='msg '+(m.from===myName?'me':'them');
    var who=document.createElement('span'); who.className='who'; who.textContent=m.from; if(m.verified){ var vb=document.createElement('span'); vb.className='vbadge'; vb.textContent='verified'; who.appendChild(vb); } if(m.admin){ var ab=document.createElement('span'); ab.className='vbadge adm'; ab.textContent='admin'; who.appendChild(ab); }
    var txt=document.createElement('span'); txt.className='txt'; txt.textContent=m.text;
    var t=document.createElement('span'); t.className='t'; t.textContent=m.time||'';
    div.appendChild(who); div.appendChild(txt); div.appendChild(t); if(m.id){ div.dataset.msgid = m.id; div.style.position='relative'; div.onclick = function(e){ if(!e.target.classList.contains('btn-del')){ showPicker(m.id, div); } }; if(m.from === myName){ var db = document.createElement('button'); db.className='btn-del'; db.innerHTML='\uD83D\uDDD1\uFE0F'; db.title='Delete'; db.onclick = (function(mid){ return function(ev){ ev.stopPropagation(); if(confirm('Hapus?')){ ws.send(JSON.stringify({type:'chat', text:'/delete '+mid})); } }; })(m.id); who.appendChild(db); } }
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
  if(dmTarget){ ws.send(JSON.stringify({type:'pm', to: dmTarget, text: txt})); } else { ws.send(JSON.stringify({type:"chat", text: txt})); }
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
		chat.SeedMsgCounter()
	}

	r := core.New()
r.Use(middleware.Recovery())
r.Use(middleware.CORS())

r.GET("/health", func(c *core.Context) {
c.JSON(http.StatusOK, core.H{"status": "ok", "app": "nexchat", "version": "1.9.0"})
})

r.GET("/", func(c *core.Context) {
c.SetHeader("Content-Type", "text/html; charset=utf-8")
c.SetHeader("Cache-Control", "no-cache")
c.Writer.WriteHeader(http.StatusOK)
c.Writer.Write([]byte(chatUI))
})

r.GET("/rooms", func(c *core.Context) {
c.JSON(http.StatusOK, core.H{"success": true, "data": hub.RoomsFull()})
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
log.Printf("[nexchat] v1.9.0 starting on :%s (powered by nexrouter!)", port)
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
	done := make(chan struct{})
verified := false
	isAdmin := false
	if tok := strings.TrimSpace(c.Query("token")); tok != "" {
		if claims, err := auth.ValidateToken(tok, jwtSecret); err == nil {
			verified = true
			if claims.Name != "" {
				name = claims.Name
			}
			isAdmin = isAdminEmail(claims.Email)
			log.Printf("[nexchat] verified user: %s (%s)", name, claims.Email)
		} else {
			log.Printf("[nexchat] invalid token: %v", err)
		}
	}
	client := hub.Join(name, room, send)
	client.Verified = verified
	client.IsAdmin = isAdmin
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
if bytes.HasPrefix(msg, []byte("__KICK__")) {
	reason := string(msg[8:])
	kb, _ := json.Marshal(map[string]string{"type": "kick", "reason": reason})
	conn.WriteMessage(websocket.TextMessage, kb)
	conn.Close()
	return
	}
	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
conn.Close()
return
}
case <-done:
			conn.Close()
			return
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
	spamGuard.Cleanup(client)
close(done)
log.Printf("[nexchat] %s left #%s", name, room)
}
func handleClientMessage(c *chat.Client, raw string) {
var env struct {
Type string `json:"type"`
Text string `json:"text"`
		MsgID int64 `json:"msg_id"`
		Emoji string `json:"emoji"`
		To    string `json:"to"`
}
if err := json.Unmarshal([]byte(raw), &env); err == nil {
if env.Type == "typing" {
hub.BroadcastTyping(c)
return
}
if env.Type == "pm" {
			if ok, sec := spamGuard.Allow(c); !ok {
				hub.NotifyClient(c, "slow down! muted "+strconv.Itoa(sec)+"s (anti-spam)")
				return
			}
			to := strings.TrimSpace(env.To)
			text := strings.TrimSpace(env.Text)
			if to == "" || text == "" {
				return
			}
			if to == c.Name {
				hub.NotifyClient(c, "cannot PM yourself")
				return
			}
			if len(text) > 500 {
				text = text[:500]
			}
			if !hub.SendPM(c, to, text) {
				hub.NotifyClient(c, "saved - "+to+" is offline; they can read via /history")
			}
			return
		}
		if env.Type == "dm_list" {
			threads := chat.DMThreads(c.Name)
			payload := map[string]interface{}{"type": "dm_list", "threads": threads}
			if b, err := json.Marshal(payload); err == nil {
				select {
				case c.Send <- b:
				default:
				}
			}
			return
		}
		if env.Type == "react" {
			if env.MsgID > 0 && env.Emoji != "" {
				snap := chat.AddReaction(env.MsgID, env.Emoji, c.Name)
				hub.BroadcastReaction(c.Room, env.MsgID, snap)
			}
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
if ok, mutedSec := spamGuard.Allow(c); !ok {
		hub.NotifyClient(c, "slow down! muted "+strconv.Itoa(mutedSec)+"s (anti-spam)")
		return
	}
	if strings.HasPrefix(raw, "/unmute ") {
		if !c.IsAdmin {
			hub.NotifyClient(c, "only admin can unmute")
			return
		}
		targetName := strings.Trim(strings.TrimSpace(raw[8:]), "\"")
		target := hub.FindClient(targetName)
		if target == nil {
			hub.NotifyClient(c, "user not found or offline: "+targetName)
			return
		}
		spamGuard.Unmute(target)
		hub.NotifyClient(c, targetName+" has been unmuted")
		hub.BroadcastSystem(c.Room, targetName+" was unmuted by "+c.Name)
		return
	}
	if strings.HasPrefix(raw, "/delete ") {
		idStr := strings.TrimSpace(raw[8:])
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			hub.NotifyClient(c, "invalid message ID")
			return
		}
		if chat.DeleteMessage(id) {
			payload := map[string]interface{}{"type": "msg_deleted", "id": id}
			if _,  err := json.Marshal(payload); err == nil {
				// Using BroadcastJSON
				
			}
		} else {
			hub.NotifyClient(c, "message not found")
		}
		return
	}
	if strings.HasPrefix(raw, "/history ") {
		target := strings.Trim(strings.TrimSpace(raw[9:]), "\"")
		if target == "" {
			hub.NotifyClient(c, "usage: /history username")
			return
		}
		msgs := chat.DMHistory(c.Name, target, 50)
		payload := map[string]interface{}{"type": "dm_history", "with": target, "messages": msgs}
		if b, err := json.Marshal(payload); err == nil {
			select {
			case c.Send <- b:
			default:
			}
		}
		return
	}
	if strings.HasPrefix(raw, "/pm ") {
		toName, pmText := parsePM(raw[4:])
		if toName != "" && pmText != "" {
			if toName == c.Name {
				hub.NotifyClient(c, "cannot PM yourself")
				return
			}
			if hub.SendPM(c, toName, pmText) {
				return
			}
			hub.NotifyClient(c, "saved - "+toName+" sedang offline; dia bisa baca via /history")
			return
		}
		hub.NotifyClient(c, "usage: /pm \"username\" message")
		return
	}
	if strings.HasPrefix(raw, "/kick ") {
		target := strings.Trim(strings.TrimSpace(raw[6:]), "\"")
		if !c.IsAdmin {
			hub.NotifyClient(c, "only admin can use /kick")
			return
		}
		if target == c.Name {
			hub.NotifyClient(c, "cannot kick yourself")
			return
		}
		if hub.Kick(target, "kicked by "+c.Name) {
			hub.BroadcastSystem(c.Room, target+" was kicked by "+c.Name)
		} else {
			hub.NotifyClient(c, "user not found: "+target)
		}
		return
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
func isAdminEmail(email string) bool {
list := os.Getenv("ADMIN_EMAILS")
if list == "" {
return false
}
for _, a := range strings.Split(list, ",") {
if strings.EqualFold(strings.TrimSpace(a), email) {
return true
}
}
return false
}

// parsePM: support  "John Doe" hello  maupun  Budi hello
func parsePM(rest string) (string, string) {
rest = strings.TrimSpace(rest)
if strings.HasPrefix(rest, "\"") {
end := strings.Index(rest[1:], "\"")
if end > 0 {
return rest[1 : 1+end], strings.TrimSpace(rest[end+2:])
}
return "", ""
}
parts := strings.SplitN(rest, " ", 2)
if len(parts) == 2 {
return parts[0], strings.TrimSpace(parts[1])
}
return "", ""
}
