document.addEventListener("DOMContentLoaded", function() {
        fetch("/feedback")
        .then(response => response.json())
        .then(data => {
            const feedbacksDiv = document.getElementById("feedbacks");
            let isArrary = Array.isArray(data)
            console.log('Fetched data:', data); // 打印数据
            for (const feedback of data) {
                const feedbackDiv = document.createElement("div");
                feedbackDiv.className = "feedback";

                const title = document.createElement("h2");
                title.textContent = feedback.Title;
                feedbackDiv.appendChild(title);

                const content = document.createElement("p");
                content.textContent = feedback.Content;
                feedbackDiv.appendChild(content);

                const filesList = document.createElement("ul");
                for (const file of feedback.FileInfos) {
                    const fileItem = document.createElement("li");
                    fileItem.textContent = `${file.OriginalName} (${file.FileType}) - ${file.FileSize} bytes`;

                    // 创建下载链接
                    const downloadLink = document.createElement("a");
                    downloadLink.href = `/downloadFile?file=${encodeURIComponent(file.ServerPath)}`; // 设置下载链接
                    downloadLink.textContent = "下载"; // 下载链接的文本
                    downloadLink.target = "_blank"; // 在新标签页中打开
                    fileItem.appendChild(downloadLink);

                    filesList.appendChild(fileItem);
                };
                feedbackDiv.appendChild(filesList);

                feedbacksDiv.appendChild(feedbackDiv);
            };
        })
        .catch(error => console.error('Error fetching feedbacks:', error));
});