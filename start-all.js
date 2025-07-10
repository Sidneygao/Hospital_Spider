const { spawn } = require('child_process');
const path = require('path');

// 启动后端
const backend = spawn('go', ['run', 'main.go', 'spider.go', 'algorithm.go'], {
  cwd: path.join(__dirname, 'backend'),
  stdio: 'inherit',
  shell: true
});

// 启动前端
const frontend = spawn('npm', ['start'], {
  cwd: path.join(__dirname, 'frontend'),
  stdio: 'inherit',
  shell: true
});

backend.on('close', code => {
  console.log(`后端进程退出，code=${code}`);
});
frontend.on('close', code => {
  console.log(`前端进程退出，code=${code}`);
}); 