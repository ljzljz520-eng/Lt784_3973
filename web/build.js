const fs = require('fs');
const path = require('path');

const output = path.join(__dirname, 'dist');
fs.mkdirSync(output, { recursive: true });
const html = `<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>权限授权对象选择器</title><link rel="stylesheet" href="app.css"></head><body><main><h1>权限授权对象选择器</h1><p>从组织树选择部门或人员，确认后提交授权。</p><section id="app"><p>服务端 API: /api/tree</p></section></main><script src="app.js"></script></body></html>`;
const css = 'body{font-family:system-ui,sans-serif;max-width:960px;margin:3rem auto;padding:0 1rem;color:#1f2937}main{border:1px solid #d1d5db;padding:2rem}h1{margin-top:0;color:#0f766e}';
const js = 'fetch("/api/tree").then(r=>r.json()).then(data=>{document.querySelector("#app").dataset.nodes=(data.nodes||[]).length}).catch(()=>{});';
fs.writeFileSync(path.join(output, 'index.html'), html);
fs.writeFileSync(path.join(output, 'app.css'), css);
fs.writeFileSync(path.join(output, 'app.js'), js);
console.log(`built ${output}`);
