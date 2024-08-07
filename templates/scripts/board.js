// 获取当前页面的URL
const currentUrl = window.location.href;

// 获取传入的Board参数
console.log(board);
window.board = board

// 生成二维码并展示在页面上
let qrcodeUrl = currentUrl
if (new URL(window.location.href).pathname === "/") {
    qrcodeUrl = currentUrl + board
}
const qrcode = new QRCode(document.getElementById("qrcode"), {
    text: qrcodeUrl, width: 128, height: 128
});


NProgress.configure({showSpinner: false});

document.addEventListener('DOMContentLoaded', (event) => {
    fetchMessages();
});

document.getElementById('messageInput').addEventListener('keydown', function (event) {
    if (event.key === 'Enter' && !event.shiftKey) {
        event.preventDefault();
        addMessage();
    }
});

document.getElementById('messageInput').addEventListener('paste', function (event) {
    const items = (event.clipboardData || event.originalEvent.clipboardData).items;
    const maxSize = 20 * 1024 * 1024; // 20MB

    const handleFile = (file) => {
        if (file.size > maxSize) {
            alert('The file size exceeds the 20MB limit.');
            return;
        }
        const reader = new FileReader();
        reader.onload = (e) => {
            const base64Image = e.target.result;
            const message = `${file.name}#${base64Image}`;
            sendMessage(message);
        };
        reader.readAsDataURL(file);
    };

    const handleText = (item) => {
        item.getAsString((text) => {
            if (new Blob([text]).size > maxSize) {
                alert('The text size exceeds the 20MB limit.');
                return;
            }
            sendMessage(text);
        });
    };

    for (const item of items) {
        // alert(`item.kind: ${item.kind} item.type: ${item.type}`);
        // if (item.kind === 'file' && item.type.startsWith('image/')) {
        if (item.kind === 'file') {
            handleFile(item.getAsFile());
            event.preventDefault();
        } else if (item.kind === 'string' && item.type === 'text/plain') {
            handleText(item);
            event.preventDefault();
        }
    }
});

function fetchMessages() {
    NProgress.start();
    fetch('/boardapi/' + board)
        .then(response => response.json())
        .then(data => {
            if (data.code == 200) {
                updateCountdown(data.data.expireAt)
                const messages = data.data.messages;
                const messageList = document.getElementById('messages');
                messageList.innerHTML = '';  // 清空现有的消息

                if (messages && messages.length == 0) {
                    const messageList = document.getElementById('messages');
                    const alertMessage = document.createElement('div');
                    alertMessage.className = 'alert alert-primary';
                    if (language == 'zh') {
                        alertMessage.innerText = '暂无内容';
                    } else {
                        alertMessage.innerText = 'No items now';
                    }
                    alertMessage.setAttribute("data-i18n", "no-messages");
                    alertMessage.style = 'margin-top: 10px;';
                    alertMessage.id = 'alert-message';
                    messageList.appendChild(alertMessage);
                } else {
                    messages.forEach(message => {
                        appendMessage(message, false);
                    });
                }
            } else {
                console.error('Error fetching messages:', data);
                alert(data.message);
                const currentUrl = new URL(window.location.href);
                const newPath = currentUrl.pathname.replace(/[^/]+$/, "public");
                window.location.href = `${currentUrl.origin}${newPath}`;
            }
        })
        .catch(error => console.error('Error fetching messages:', error))
        .finally(() => {
            // Complete the progress bar
            NProgress.done();
        });
}

function addMessage() {
    const input = document.getElementById('messageInput');
    const messageText = input.value.trim();
    if (messageText) {
        sendMessage(messageText);
        input.value = '';
    }
}

function sendMessage(content) {
    NProgress.start();
    fetch('/boardapi/' + board, {
        method: 'POST', headers: {
            'Content-Type': 'application/json'
        }, body: JSON.stringify({content: content})
    }).then(response => response.json())
        .then(data => {
            if (data.code == 200 && data.data.messages && data.data.messages.length > 0) {
                updateCountdown(data.data.expireAt)
                appendMessage(data.data.messages[0], true);
                snapdrop.send({type: 'board-update', board: window.board})
                if (language == 'zh') {
                    Events.fire('notify-user', '剪贴板记录添加成功');
                } else {
                    Events.fire('notify-user', 'Clipboard record added successfully');
                }
            } else {
                console.error('Error adding message:', data);
                alert(data.message);
            }
        })
        .catch(error => {
            console.error('Error adding message:', error);
            alert('Error adding message: ' + error);
        })
        .finally(() => {
            // Complete the progress bar
            NProgress.done();
        });
}

function appendMessage(message, appendMessageToFirst) {
    // messageList 中如果有div的Element，全部删除
    const alertEle = document.getElementById('alert-message');
    if (alertEle != null) {
        alertEle.remove();
    }

    const messageList = document.getElementById('messages');
    const newMessage = document.createElement('li');
    newMessage.setAttribute('data-id', message.id);

    // console.log("message.content: " + message.content);

    const isFile = message.isFile;

    if (isFile && message.fileType.startsWith('image/')) {
        const img = document.createElement('img');
        img.src = `/boardapi/${board}/${message.id}`;
        newMessage.appendChild(img);

        img.addEventListener('click', () => {
            const overlay = document.createElement('div');
            overlay.classList.add('overlay');

            const largeImage = document.createElement('img');
            largeImage.src = img.src;
            largeImage.classList.add('expanded');

            overlay.appendChild(largeImage);
            document.body.appendChild(overlay);

            setTimeout(() => {
                overlay.classList.add('active');
            }, 10);

            overlay.addEventListener('click', () => {
                overlay.classList.remove('active');
                largeImage.classList.remove('expanded');
                setTimeout(() => overlay.remove(), 300);
            });
        });

    } else if (isFile) {
        // 其他文件
        const filenameDisplay = document.createElement('a');
        filenameDisplay.href = `/boardapi/${board}/${message.id}`;
        filenameDisplay.innerText = message.fileName;
        filenameDisplay.download = message.fileName;
        newMessage.appendChild(filenameDisplay);
    } else {
        if (isValidURL(message.content)) {
            const link = document.createElement('a');
            link.href = message.content;
            link.textContent = message.content;
            link.target = '_blank';
            newMessage.appendChild(link);
        } else {
            newMessage.textContent = message.content;
        }
    }

    const buttonsDiv = document.createElement('div');
    buttonsDiv.className = 'buttons';

    if (!message.isFile || message.fileType == 'image/png') {
        const copyButton = document.createElement('button');
        copyButton.className = 'btn btn-primary btn-sm';
        copyButton.title = "复制";
        const copyIcon = document.createElement('i');
        copyIcon.className = 'iconfont icon-copy';
        copyButton.addEventListener('click', () => copyToClipboard(message));
        copyButton.appendChild(copyIcon);
        buttonsDiv.appendChild(copyButton);
    }

    const deleteButton = document.createElement('button');
    deleteButton.className = 'btn btn-secondary btn-sm';
    deleteButton.title = "删除";
    const deleteIcon = document.createElement('i');
    deleteIcon.className = 'iconfont icon-remove';
    deleteButton.addEventListener('click', () => deleteMessage(message.id, newMessage));
    deleteButton.appendChild(deleteIcon);
    buttonsDiv.appendChild(deleteButton);

    newMessage.appendChild(buttonsDiv);

    if (appendMessageToFirst) {
        messageList.insertBefore(newMessage, messageList.firstChild);
    } else {
        messageList.appendChild(newMessage);
    }

    // 保证 messageList 最多只保留前面 5 条
    while (messageList.children.length > 5) {
        messageList.removeChild(messageList.lastChild);
    }
}

function copyToClipboard(message) {
    if (message.isFile && message.fileType == 'image/png') {
        const url = `/boardapi/${board}/${message.id}`
        fetch(url).then(res => {
            if (!res.ok) {
                throw new Error(`Network response was not ok: ${res.statusText}`);
            }
            return res.blob();
        }).then(blob => {
            const item = new ClipboardItem({'image/png': blob});
            navigator.clipboard.write([item]).then(() => {
                if (language == 'zh') {
                    Events.fire('notify-user', '已复制到剪贴板');
                } else {
                    Events.fire('notify-user', 'Copied to clipboard');
                }
            }).catch(error => {
                console.error('Error copying to clipboard:', error);
            });
        }).catch(error => {
            console.error('Failed to get messege:', error);
        });
    } else if (!message.isFile) {
        const tempInput = document.createElement('textarea');
        tempInput.value = message.content;
        document.body.appendChild(tempInput);
        tempInput.select();
        document.execCommand('copy');
        document.body.removeChild(tempInput);
        if (language == 'zh') {
            Events.fire('notify-user', '已复制到剪贴板');
        } else {
            Events.fire('notify-user', 'Copied to clipboard');
        }
    } else {
        if (language == 'zh') {
            Events.fire('notify-user', '不支持的文件格式');
        } else {
            Events.fire('notify-user', 'Not supported');
        }
    }
}

function isValidURL(str) {
    const pattern = /^(https?:\/\/)[^\s/$.?#].[^\s]*$/i;
    return pattern.test(str);
}

function copyUrlToClipboard() {
    const urlText = document.getElementById('url').innerText;
    navigator.clipboard.writeText(urlText).then(function () {
        if (language == 'zh') {
            Events.fire('notify-user', '已复制到剪贴板');
        } else {
            Events.fire('notify-user', 'Copied to clipboard');
        }
    }, function (err) {
        console.error('无法复制内容：', err);
    });
}

function deleteMessage(id, messageElement) {
    NProgress.start();
    fetch(`/boardapi/` + board + `/${id}`, {
        method: 'DELETE'
    })
        .then(response => response.json())
        .then(data => {
            if (data.code == 200) {
                messageElement.remove();

                if (data.data.messages && data.data.messages.length == 0) {
                    const messageList = document.getElementById('messages');
                    messageList.innerHTML = '';  // 清空现有的消息
                    const alertMessage = document.createElement('div');
                    alertMessage.className = 'alert alert-primary';
                    if (language == 'zh') {
                        alertMessage.innerText = '暂无内容';
                    } else {
                        alertMessage.innerText = 'No items now';
                    }
                    alertMessage.setAttribute("data-i18n", "no-messages");
                    alertMessage.style = 'margin-top: 10px;';
                    alertMessage.id = 'alert-message';
                    messageList.appendChild(alertMessage);
                }

                updateCountdown(data.data.expireAt)

                snapdrop.send({type: 'board-update', board: window.board})
                if (language == 'zh') {
                    Events.fire('notify-user', '剪贴板记录删除成功');
                } else {
                    Events.fire('notify-user', 'Clipboard record deleted');
                }
            } else {
                console.error('Error deleting message:', data);
            }
        })
        .catch(error => console.error('Error deleting message:', error))
        .finally(() => {
            // Complete the progress bar
            NProgress.done();
        });
}

let translations_zh = {};
let translations_en = {};

let language = 'zh';

function loadTranslations(callback) {
    const loadJSON = (lang) => {
        return fetch(`/css/board-${lang}.json`)
            .then(response => response.json())
            .then(data => {
                if (lang === 'en') {
                    translations_en = data;
                } else if (lang === 'zh') {
                    translations_zh = data;
                }
            })
            .catch(error => console.error(`Error loading ${lang} translation file:`, error));
    };

    Promise.all([loadJSON('en'), loadJSON('zh')])
        .then(() => {
            if (callback) {
                callback();
            }
        });
}

function switchLanguage(lang) {
    document.querySelectorAll('[data-i18n]').forEach(el => {
        const key = el.getAttribute('data-i18n');
        if (lang == 'en') {
            if (translations_en[key]) {
                el.textContent = translations_en[key];
            }
        } else {
            if (translations_zh[key]) {
                el.textContent = translations_zh[key];
            }
        }
    });

    document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
        const key = el.getAttribute('data-i18n-placeholder');
        if (lang == 'en') {
            if (translations_en[key]) {
                el.setAttribute('placeholder', translations_en[key]);
            }
        } else {
            if (translations_zh[key]) {
                el.setAttribute('placeholder', translations_zh[key]);
            }
        }
    });

    document.querySelectorAll('x-instructions').forEach(el => {
        if (lang == 'en') {
            if (translations_en["desktop-instructions"]) {
                el.setAttribute('desktop', translations_en["desktop-instructions"]);
            }
            if (translations_en["mobile-instructions"]) {
                el.setAttribute('mobile', translations_en["mobile-instructions"]);
            }
        } else {
            if (translations_zh["desktop-instructions"]) {
                el.setAttribute('desktop', translations_zh["desktop-instructions"]);
            }
            if (translations_zh["mobile-instructions"]) {
                el.setAttribute('mobile', translations_zh["mobile-instructions"]);
            }
        }
    });


}

function switchLanguageOnload(lang) {
    loadTranslations(() => {
        switchLanguage(lang);
    });
}

// 获取倒计时显示元素
const countdownElement = document.getElementById('countdown');// 设置提示文本
countdownElement.title = "到期将自动清理剪切板";

// 函数：更新倒计时
let interval;

function updateCountdown(expireAt) {
    // 清除之前的倒计时
    if (interval) {
        clearInterval(interval);
    }

    // 将过期时间转换为时间戳
    const expireTime = new Date(expireAt).getTime();

    // 更新倒计时的内部函数
    function countdown() {
        const now = new Date().getTime();
        const distance = expireTime - now;

        // 计算天、小时、分钟和秒
        const hours = Math.floor((distance % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
        const minutes = Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60));
        const seconds = Math.floor((distance % (1000 * 60)) / 1000);

        // 补足两位数
        const hoursStr = String(hours).padStart(2, '0');
        const minutesStr = String(minutes).padStart(2, '0');
        const secondsStr = String(seconds).padStart(2, '0');

        // 更新倒计时显示内容
        countdownElement.innerHTML = `
            <span class="countdown-part">${hoursStr}</span>
            <span class="countdown-separator">:</span>
            <span class="countdown-part">${minutesStr}</span>
            <span class="countdown-separator">:</span>
            <span class="countdown-part">${secondsStr}</span>
        `;

        // 如果倒计时结束，停止更新
        if (distance < 0) {
            clearInterval(interval);
            countdownElement.innerText = "已过期";
        }
    }

    // 每秒更新一次倒计时
    interval = setInterval(countdown, 1000);

    // 立即调用一次以显示初始倒计时
    countdown();
}

document.addEventListener("DOMContentLoaded", function () {
    const boardInput = document.getElementById("board-input");

    boardInput.addEventListener("click", function () {
        boardInput.select();
    });

    const tooltipIcon = document.getElementById('tooltip-icon');
    if (board === "public") {
        tooltipIcon.style.visibility = "visible";

        const tipSpan = document.createElement('span');
        tipSpan.className = 'tooltip-text';
        tipSpan.setAttribute("data-i18n", "tooltip");
        if (language == 'en') {
            tipSpan.textContent = 'This is the public clipboard space. You can customize the clipboard space in the blue box.';
        } else {
            tipSpan.textContent = '这里是公共剪贴板空间 public，您可以在蓝色框内自定义剪贴板空间。';
        }

        tooltipIcon.appendChild(tipSpan)
    } else {
        tooltipIcon.style.visibility = "hidden";
    }
});

function handleKeyDown(event) {
    if (event.key === "Enter") {
        navigateToBoard();
    }
}

function handleBlur() {
    navigateToBoard();
}

function navigateToBoard() {
    const boardInput = document.getElementById("board-input");
    const newBoard = boardInput.value.trim();
    if (newBoard && newBoard !== board) {
        const currentUrl = new URL(window.location.href);
        if (currentUrl.pathname === "/") {
            window.location.href = `${currentUrl.origin}/${newBoard}`;
        } else {
            const newPath = currentUrl.pathname.replace(/[^/]+$/, newBoard);
            window.location.href = `${currentUrl.origin}${newPath}`;
        }
    }
}