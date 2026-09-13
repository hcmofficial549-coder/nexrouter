$mainPath = "$PWD\cmd\chat\main.go"
$main = [System.IO.File]::ReadAllText($mainPath)

# 1. Tambah myAdmin
if (-not $main.Contains("var myAdmin")) {
    $main = $main.Replace("var dmTarget = null;", "var dmTarget = null;" + "`r`n" + "var myAdmin = false;")
    Write-Host "[1/4] OK: myAdmin added" -ForegroundColor Green
}

# 2. Tambah removeMsg function
if (-not $main.Contains("function removeMsg")) {
    # Kita pakai single quote untuk JS literal agar aman
    $rm = 'function removeMsg(id){' + "`r`n"
    $rm += '  var els = document.querySelectorAll(''[data-msgid="'' + id + ''"]'');' + "`r`n"
    $rm += '  for(var i=0; i<els.length; i++){' + "`r`n"
    $rm += '    els[i].style.opacity = ''0.3'';' + "`r`n"
    $rm += '    els[i].style.textDecoration = ''line-through'';' + "`r`n"
    $rm += '    setTimeout((function(el){ return function(){ el.remove(); }; })(els[i]), 500);' + "`r`n"
    $rm += '  }' + "`r`n"
    $rm += '}' + "`r`n"
    $main = $main.Replace("function addMsg(m){", $rm + "`r`n" + "function addMsg(m){")
    Write-Host "[2/4] OK: removeMsg added" -ForegroundColor Green
}

# 3. Tambah CSS btn-del
if (-not $main.Contains(".btn-del")) {
    $css = '.btn-del{background:none;border:none;color:#ef4444;cursor:pointer;font-size:.7rem;padding:0 4px;opacity:0.6;margin-left:4px}' + "`r`n"
    $css += '.btn-del:hover{opacity:1}' + "`r`n"
    $main = $main.Replace('.msg .who{', $css + '.msg .who{')
    Write-Host "[3/4] OK: CSS btn-del added" -ForegroundColor Green
}

# 4. Tambah tombol delete di bubble
$old = 'div.onclick = function(){ showPicker(m.id, div); }; }'
$new = 'div.onclick = function(e){ if(!e.target.classList.contains(''btn-del'')){ showPicker(m.id, div); } }; if(m.from === myName){ var db = document.createElement(''button''); db.className=''btn-del''; db.innerHTML=''\uD83D\uDDD1\uFE0F''; db.title=''Delete''; db.onclick = (function(mid){ return function(ev){ ev.stopPropagation(); if(confirm(''Hapus?'')){ ws.send(JSON.stringify({type:''chat'', text:''/delete ''+mid})); } }; })(m.id); who.appendChild(db); } }'

if ($main.Contains($old)) {
    $main = $main.Replace($old, $new)
    Write-Host "[4/4] OK: Delete button added" -ForegroundColor Green
} else {
    Write-Host "[4/4] WARNING: Anchor bubble tidak ditemukan." -ForegroundColor Yellow
}

# 5. Routing event msg_deleted
if (-not $main.Contains("msg_deleted")) {
    $main = $main.Replace("} else if(mm.type === 'dm_list'){", "} else if(mm.type === 'msg_deleted'){ removeMsg(mm.id); } else if(mm.type === 'dm_list'){")
    Write-Host "[5/5] OK: msg_deleted routing added" -ForegroundColor Green
}

[System.IO.File]::WriteAllText($mainPath, $main)

Write-Host ""
Write-Host "=== Verifikasi ===" -ForegroundColor Cyan
$chk = [System.IO.File]::ReadAllText($mainPath)
$ok = 0
foreach ($n in @("btn-del", "removeMsg", "msg_deleted", "myAdmin")) {
    if ($chk.Contains($n)) { Write-Host "  [OK] $n" -ForegroundColor Green; $ok++ }
    else { Write-Host "  [MISSING] $n" -ForegroundColor Red }
}
if ($ok -eq 4) { Write-Host "SEMUA LENGKAP! Lanjut Build." -ForegroundColor Green }