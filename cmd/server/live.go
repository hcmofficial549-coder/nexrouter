package main

import (
"encoding/json"
"fmt"
"net/http"
"time"

"github.com/nexrouter/nexrouter/core"
)

const monitorHTML = `<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>nexrouter Live Monitor</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{background:#0a0a0f;color:#e4e4e7;font-family:'Courier New',monospace;padding:20px}
h1{font-size:18px;margin-bottom:2px;color:#06b6d4}
.sub{color:#71717a;font-size:12px;margin-bottom:14px}
.bar{display:flex;gap:10px;align-items:center;margin-bottom:14px;flex-wrap:wrap}
.pill{padding:4px 11px;border-radius:999px;font-size:11px;background:#1a1a2e;border:1px solid #27272a}
.dot{width:8px;height:8px;border-radius:50%;background:#ef4444;display:inline-block;margin-right:6px}
.dot.on{background:#10b981;box-shadow:0 0 8px #10b981}
button{background:#1a1a2e;border:1px solid #27272a;color:#a1a1aa;padding:5px 12px;border-radius:6px;cursor:pointer;font-family:inherit;font-size:12px}
button:hover{border-color:#06b6d4;color:#06b6d4}
#feed{display:flex;flex-direction:column;gap:4px}
.ev{display:flex;gap:10px;padding:7px 10px;background:#12121a;border:1px solid #27272a;border-radius:6px;font-size:12px;align-items:center}
.ev .t{color:#71717a;min-width:62px}
.ev .m{font-weight:bold;min-width:56px}
.mGET{color:#10b981}.mPOST{color:#3b82f6}.mPUT{color:#f59e0b}.mDELETE{color:#ef4444}
.ev .p{color:#06b6d4;flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.s{font-weight:bold;min-width:34px;text-align:right}
.sok{color:#10b981}.serr{color:#ef4444}
.l{color:#71717a;min-width:70px;text-align:right}
.empty{color:#71717a;text-align:center;padding:40px;font-size:13px;line-height:1.8}
</style>
</head>
<body>
<h1>nexrouter LIVE MONITOR</h1>
<div class="sub">Realtime request feed - Server-Sent Events (SSE) - auto-reconnect</div>
<div class="bar">
<span class="pill"><span class="dot" id="dot"></span><span id="conn">connecting...</span></span>
<span class="pill">events: <b id="cnt">0</b></span>
<span class="pill">errors: <b id="errcnt" style="color:#ef4444">0</b></span>
<button onclick="clearFeed()">clear</button>
</div>
<div id="feed"><div class="empty">Menunggu request...<br>Buka /dashboard atau API di tab lain, klik-klik -<br>event akan muncul DI SINI secara realtime!</div></div>
<script>
var cnt=0, errs=0;
var es = new EventSource("/api/v1/events/stream");
es.onopen = function(){ document.getElementById("dot").className="dot on"; document.getElementById("conn").textContent="LIVE"; };
es.onerror = function(){ document.getElementById("dot").className="dot"; document.getElementById("conn").textContent="reconnecting..."; };
es.onmessage = function(e){
  var ev = JSON.parse(e.data);
  cnt++;
  document.getElementById("cnt").textContent=cnt;
  if(ev.status>=400){ errs++; document.getElementById("errcnt").textContent=errs; }
  var feed=document.getElementById("feed");
  var em=feed.querySelector(".empty");
  if(em){ em.remove(); }
  var div=document.createElement("div");
  div.className="ev";
  var sc = ev.status<400 ? "sok" : "serr";
  div.innerHTML='<span class="t">'+ev.time+'</span><span class="m m'+ev.method+'">'+ev.method+'</span><span class="p">'+ev.path+'</span><span class="s '+sc+'">'+ev.status+'</span><span class="l">'+ev.latency_ms+' ms</span>';
  feed.insertBefore(div, feed.firstChild);
  while(feed.children.length>200){ feed.removeChild(feed.lastChild); }
};
function clearFeed(){
  document.getElementById("feed").innerHTML='<div class="empty">Feed cleared - menunggu request...</div>';
  cnt=0; errs=0;
  document.getElementById("cnt").textContent="0";
  document.getElementById("errcnt").textContent="0";
}
</script>
</body>
</html>
`

// monitorHandler serves the live monitor page
func monitorHandler(c *core.Context) {
c.SetHeader("Content-Type", "text/html; charset=utf-8")
c.SetHeader("Cache-Control", "no-cache")
c.Writer.WriteHeader(http.StatusOK)
c.Writer.Write([]byte(monitorHTML))
}

// eventsStreamHandler streams events via Server-Sent Events
func eventsStreamHandler(c *core.Context) {
flusher, ok := c.Writer.(http.Flusher)
if !ok {
c.InternalError("streaming unsupported")
return
}

ch, unsub := eventHub.Subscribe()
defer unsub()

c.SetHeader("Content-Type", "text/event-stream")
c.SetHeader("Cache-Control", "no-cache")
c.SetHeader("Connection", "keep-alive")
c.SetHeader("X-Accel-Buffering", "no")
c.Writer.WriteHeader(http.StatusOK)
flusher.Flush()

ctx := c.Request.Context()
heartbeat := time.NewTicker(15 * time.Second)
defer heartbeat.Stop()

for {
select {
case <-ctx.Done():
return
case ev := <-ch:
data, err := json.Marshal(ev)
if err != nil {
continue
}
fmt.Fprintf(c.Writer, "data: %s\n\n", data)
flusher.Flush()
case <-heartbeat.C:
fmt.Fprint(c.Writer, ": ping\n\n")
flusher.Flush()
}
}
}