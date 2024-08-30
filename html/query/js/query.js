// 分页大小
const PageSize = 10;

// 监听页面加载事件
document.addEventListener('DOMContentLoaded', function() {  
    fetchData(0, PageSize);
});  
  
// 请求数据
function fetchData(pageIndex, pageSize) {  
    // 准备JSON数据  
    const postData = JSON.stringify({  
        pageIndex: pageIndex,  
        pageSize: pageSize  
    });  
  
    // 配置fetch请求  
    const options = {  
        method: 'POST',  
        headers: {  
            'Content-Type': 'application/json',  
        },  
        body: postData  
    };  
  
    // 执行fetch 
    fetch('/api/queryFeedback', options) 
        .then(response => {  
            if (!response.ok) {  
                throw new Error('Network response was not ok');  
            }  
            return response.json();  
        })  
        .then(data => {  
            // 渲染返回结果  
            renderFeedbackList(data); 

            // 渲染分页按钮  
            const paginationDiv = document.getElementById('pagination');  
            paginationDiv.innerHTML = ''; // 清空之前的分页按钮  

            if (data.totalSize === 0) {
                return
            }

            // 向上取整分页数量
            var totalPages = Math.ceil(data.totalSize / PageSize);
    
            // 遍历页码并添加按钮  
            for (let i = 1; i <= totalPages; i++) {  
                const button = document.createElement('button');  
                button.textContent = i;  
                button.classList.add('page-button');  
                button.onclick = function() {  
                    fetchData(this.textContent - 1, PageSize)
                };  

                if (i - 1 === data.currentPageIndex) {  
                    button.disabled = true; // 禁用当前页码按钮（可选）  
                    button.classList.add('active'); // 添加活动类（可选）  
                }  

                paginationDiv.appendChild(button);  
            }  
        })  
        .catch(error => console.error('Error fetching data:', error));  
}  
  
// 转换level为字符串
function frequencyString(level) {
    if (level === 0) {
        return '偶尔'
    }
    else if (level === 1) {
        return '经常'
    }
    else if (level === 2) {
        return '总是'
    }
    else {
        return ''
    }
}

// 转换时间戳为字符串
function formatTimestamp(timestamp) {  
    // 创建一个Date对象  
    const date = new Date(timestamp);  

    // 使用getFullYear(), getMonth() + 1, getDate(), getHours(), getMinutes(), getSeconds() 获取年月日时分秒  
    // 注意：getMonth() 返回的是0-11，所以需要+1  
    const year = date.getUTCFullYear();  
    const month = String(date.getUTCMonth() + 1).padStart(2, '0'); // 使用padStart确保月份是两位数  
    const day = String(date.getUTCDate()).padStart(2, '0');  
    const hours = String(date.getUTCHours()).padStart(2, '0');  
    const minutes = String(date.getUTCMinutes()).padStart(2, '0');  
    const seconds = String(date.getUTCSeconds()).padStart(2, '0');  

    // 拼接成YYYY-MM-DD HH:MM:SS格式  
    return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;  
}  

// 渲染请求数据
function renderFeedbackList(feedbacks) {  
    const tbody = document.querySelector('#dataList tbody');  
    tbody.innerHTML = ''; // 清空之前的列表项（如果有的话）  
  
    feedbacks.pageData.forEach(item => {  
        const row = tbody.insertRow();  
        const cellModule = row.insertCell(0);  
        const cellFreqency = row.insertCell(1);  
        const cellDescription = row.insertCell(2);  
        const cellSteps = row.insertCell(3);
        const cellAppVersion = row.insertCell(4);
        const cellTimeStamp = row.insertCell(5);
        const cellEmail = row.insertCell(6);
        const cellFile = row.insertCell(7);
        const cellOperate = row.insertCell(8);

        // 填充数据  
        cellModule.textContent = item.impactedModule;  
        cellFreqency.textContent = frequencyString(item.occurringFrequency);  
        cellDescription.textContent = item.bugDescription;  
        cellSteps.textContent = item.reproduceSteps;
        cellAppVersion.textContent = item.appVersion;
        cellTimeStamp.textContent = formatTimestamp(item.timeStamp);
        cellEmail.textContent = item.email

        // 创建文件列表的容器  
        const fileList = document.createElement('ul');  
        // 遍历文件数组并添加下载链接  
        (item.files || []).forEach(file => {
            const fileItem = document.createElement('li');  
            const downloadLink = document.createElement('a');  
            downloadLink.href = file.filePathOnOss;
            downloadLink.textContent = file.fileName;
            downloadLink.target = '_blank'; // 新标签页打开  
            downloadLink.setAttribute('download', file.filename); 
            fileItem.appendChild(downloadLink);  
            fileList.appendChild(fileItem);  
        });  

        // 将文件列表添加到 cellFile 单元格  
        cellFile.appendChild(fileList);  

        const operationBtnDelete = document.createElement('button');  
        operationBtnDelete.textContent = '删除';  
        operationBtnDelete.onclick = function() {  
            const body = JSON.stringify({ feedbackID: [parseInt(item.feedbackID, 10)] }); // 创建请求体  
  
            // 发送POST请求  
            fetch('/api/deleteFeedback', {  
                method: 'POST',  
                headers: {  
                    'Content-Type': 'application/json' // 设置请求头  
                },  
                body: body // 设置请求体  
            })  
            .then(response => {  
                if (!response.ok) {  
                    throw new Error('Network response was not ok');  
                }  
                return response.text(); // 解析TEXT响应 
            })  
            .then(data => {  
                console.log('Success:', data); // 处理成功的情况  

                let rowCount = tbody.rows.length;
                let queryPageIndex = feedbacks.currentPageIndex;
                if (rowCount === 1) {
                    queryPageIndex -= 1;
                }
                if (queryPageIndex < 0) {
                    queryPageIndex = 0;
                }

                // 重新请求数据
                fetchData(queryPageIndex, PageSize);
            })  
            .catch(error => {  
                console.error('There has been a problem with your fetch operation:', error);  
            });  
        };  
        cellOperate.appendChild(operationBtnDelete);
    });  
}  