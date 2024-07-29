document.addEventListener('DOMContentLoaded', function() {  
    fetch('/api/feedbacks')  
        .then(response => response.json())  
        .then(data => {  
            const list = document.getElementById('feedback-list');  
            data.forEach(feedback => {  
                const item = document.createElement('li');  
                item.textContent = `ID: ${feedback.feedback_id}, Title: ${feedback.title}, Content: ${feedback.content.substring(0, 100)}...`; // 截取内容以避免过长  
  
                // 假设你想显示文件信息（这里只显示第一个文件名作为示例）  
                if (feedback.files && feedback.files.length > 0) {  
                    item.textContent += `, Files: [${feedback.files[0].file_name}]`;  
                }  
  
                list.appendChild(item);  
            });  
        })  
        .catch(error => console.error('Error fetching feedbacks:', error));  
});