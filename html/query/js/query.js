/*
 * @Author: shanghanjin
 * @Date: 2024-07-30 10:09:12
 * @LastEditTime: 2024-08-25 20:04:51
 * @FilePath: \UserFeedBack\html\query\js\query.js
 * @Description: 
 */
document.addEventListener("DOMContentLoaded", function() {  
    function fetchFeedbacks(pageIndex = 0, pageSize = 10) {  
        fetch("/api/queryFeedback", {  
            method: 'POST',  
            headers: {  
                'Content-Type': 'application/json',  
            },  
            body: JSON.stringify({ pageSize, pageIndex })  
        })  
        .then(response => response.json())  
        .then(data => {  
            const feedbacksDiv = document.getElementById("feedbacks");  
            feedbacksDiv.innerHTML = ''; // 清空之前的反馈  
  
            data.forEach(feedback => {  
                const feedbackDiv = document.createElement("div");  
                feedbackDiv.className = "feedback";  
                feedbackDiv.onclick = function() { showDetail(feedback); }; // 添加点击事件  
  
                const title = document.createElement("h3");  
                title.textContent = `${feedback.impactedModule} - ${feedback.occurringFrequency} - ${feedback.bugDescription}`;  
                feedbackDiv.appendChild(title);  
  
                feedbacksDiv.appendChild(feedbackDiv);  
            });  
  
            if (data.length < pageSize) {  
                // 可以在这里添加没有更多数据的处理  
            }  
        })  
        .catch(error => console.error('Error fetching feedbacks:', error));  
    }  
  
    function showDetail(feedback) {  
        const detailContent = document.getElementById("detailContent");  
        detailContent.innerHTML = ''; // 清空之前的详细信息  
  
        const detailDiv = document.createElement("div");  
  
        // 添加所有信息  
        Object.keys(feedback).forEach(key => {  
            if (key !== 'files') {  
                const p = document.createElement("p");  
                p.textContent = `${key}: ${feedback[key]}`;  
                detailDiv.appendChild(p);  
            }  
        });  
  
        // 添加文件下载链接  
        if (feedback.files) {  
            const filesList = document.createElement("ul");  
            feedback.files.forEach(file => {  
                const fileItem = document.createElement("li");  
                const link = document.createElement("a");  
                link.href = file.filePathOnOss;  
                link.textContent = `${file.filename} (${file.fileSize} bytes)`;  
                link.target = "_blank";  
                fileItem.appendChild(link);  
                filesList.appendChild(fileItem);  
            });  
            detailDiv.appendChild(filesList);  
        }  
  
        detailContent.appendChild(detailDiv);  
        document.getElementById("feedbackDetail").style.display = 'block';  
    }  
  
    // 初始加载第一页数据  
    fetchFeedbacks();  
});