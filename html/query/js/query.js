const PageSize = 10;

document.addEventListener('DOMContentLoaded', function() {  
    fetchData(0, PageSize);
});  
  
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
  
    fetch('/api/queryFeedback', options) // 假设这是您的API URL  
        .then(response => {  
            if (!response.ok) {  
                throw new Error('Network response was not ok');  
            }  
            return response.json();  
        })  
        .then(data => {  
            renderFeedbackList(data); // 假设API返回的数据中有一个名为pageData的数组  

            // 渲染分页按钮  
            const paginationDiv = document.getElementById('pagination');  
            paginationDiv.innerHTML = ''; // 清空之前的分页按钮  

            if (data.totalSize === 0) {
                return
            }

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

function renderFeedbackList(feedbacks) {  
    const tbody = document.querySelector('#dataList tbody');  
    tbody.innerHTML = ''; // 清空之前的列表项（如果有的话）  
  
    feedbacks.pageData.forEach(item => {  
        const row = tbody.insertRow();  
        const cellModule = row.insertCell(0);  
        const cellFreqency = row.insertCell(1);  
        const cellDescription = row.insertCell(2);  
        const cellSteps = row.insertCell(3);
        const cellEmail = row.insertCell(4);
        const cellFile = row.insertCell(5);
        const cellOperate = row.insertCell(6);

        // 填充数据  
        cellModule.textContent = item.impactedModule;  
        cellFreqency.textContent = frequencyString(item.occurringFrequency);  
        cellDescription.textContent = item.bugDescription;  
        cellSteps.textContent = item.reproduceSteps;
        cellEmail.textContent = item.email

        // 创建文件列表的容器  
        const fileList = document.createElement('ul');  
        // 遍历文件数组并添加下载链接  
        (item.files || []).forEach(file => { // 假设 item.files 是一个包含文件对象的数组  
            const fileItem = document.createElement('li');  
            const downloadLink = document.createElement('a');  
            downloadLink.href = file.filePathOnOss; // 假设每个文件对象都有一个 url 属性  
            downloadLink.textContent = file.fileName; // 假设每个文件对象都有一个 name 属性  
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
                return response.text(); // 解析JSON响应（如果有的话）  
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
                fetchData(queryPageIndex, PageSize);
            })  
            .catch(error => {  
                console.error('There has been a problem with your fetch operation:', error);  
            });  
        };  

        cellOperate.appendChild(operationBtnDelete);
    });  
}  