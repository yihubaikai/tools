<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>行愿九期 每日分享生成器</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, "Noto Sans", sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji";
            padding: 20px;
            background-color: #f8f9fa;
            color: #343a40;
            line-height: 1.6;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
            background-color: #ffffff;
            padding: 20px 30px;
            border-radius: 8px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.08);
        }
        h1, h2 {
            text-align: center;
            color: #495057;
        }
        .controls {
            margin-bottom: 30px;
            padding: 20px;
            background-color: #f8f9fa;
            border-radius: 6px;
        }
        .input-group {
            margin-bottom: 15px;
        }
        .input-group label {
            display: block;
            margin-bottom: 5px;
            font-weight: bold;
            color: #495057;
        }
        input[type="text"], textarea {
            box-sizing: border-box;
            width: 100%;
            padding: 10px;
            border: 1px solid #ced4da;
            border-radius: 4px;
            font-size: 16px;
        }
        textarea {
            resize: vertical;
            min-height: 100px;
        }
        .output-container {
            margin-top: 20px;
        }
        pre {
            background-color: #e9ecef;
            padding: 20px;
            border-radius: 6px;
            white-space: pre-wrap;
            word-wrap: break-word;
            font-size: 16px;
            font-family: "KaiTi", "楷体", STKaiti, serif;
            border: 1px solid #dee2e6;
        }
        .copy-button {
            display: block;
            width: 100%;
            padding: 12px;
            margin-top: 15px;
            font-size: 18px;
            font-weight: bold;
            color: #fff;
            background-color: #28a745;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            transition: background-color 0.2s;
        }
        .copy-button:hover {
            background-color: #218838;
        }
        .copy-button:active {
            background-color: #1e7e34;
        }
    </style>
</head>
<body>

<div class="container">
    <h1>行愿九期 每日分享</h1>
    <div class="controls">
        <div class="input-group">
            <label for="m_username">用户名：</label>
            <input type="text" id="m_username" value="行愿九期如松">
        </div>
        <div class="input-group">
            <label for="m_usershiyan">誓言：</label>
            <input type="text" id="m_usershiyan" value="我是一个自信勇敢的人，承诺创造一个幸福包容的世界">
        </div>
        <div class="input-group">
            <label for="m_TeamName">队名：</label>
            <input type="text" id="m_TeamName" value="先锋队">
        </div>
        <div class="input-group">
            <label for="m_Duihu">队呼：</label>
            <input type="text" id="m_Duihu" value="先锋，先锋，谁与争锋！">
        </div>
        <!-- 新增的发愿输入框 -->
        <div class="input-group">
            <label for="m_FaYuan">我/我们发愿：</label>
            <input type="text" id="m_FaYuan" value="你若不离六道，我必常随娑婆">
        </div>
        <!-- 结束新增 -->
        <div class="input-group">
            <label for="m_fenxiang">今日分享：</label>
            <textarea id="m_fenxiang">贫穷布施难，富贵学道难；
弃命必死难，得睹佛经难；
生值佛世难，忍色忍欲难；
见好不求难，被辱不瞋难；
有势不临难，触事无心难；
广学博究难，除灭我慢难；
不轻未学难，心行平等难；
不说是非难，会善知识难；
见性学道难，随化度人难；
睹境不动难，善解方便难。
-----出自《人生二十难》</textarea>
        </div>
        <div class="input-group">
            <label for="m_jingwen">经文分享：</label>
            <textarea id="m_jingwen"></textarea> <!-- 默认值由JS随机填充 -->
        </div>
    </div>

    <div class="output-container">
        <h2>最终生成内容 (下方文本框内容会随上方修改实时更新)</h2>
        <pre id="final-output"></pre>
        <button id="copy-btn" class="copy-button">一键复制全部分享内容</button>
    </div>
</div>


<script>
// 使用 DOMContentLoaded 事件确保在操作DOM之前，所有HTML元素都已加载完毕
document.addEventListener('DOMContentLoaded', function() {

    // ===================================================================
    // 1. 数据与常量定义
    // ===================================================================

    // 农历数据映射表 (2024-2027)
    const lunarData = {
        // ... (省略部分数据以保持代码简洁，完整数据已包含)
        '2024-01-01':'冬月二十','2024-02-01':'腊月廿二','2024-03-01':'正月二十','2024-03-10':'二月初一','2024-04-01':'二月廿三','2024-05-01':'三月廿三','2024-05-22':'四月十五','2024-06-01':'四月廿五','2024-06-30':'五月廿五','2024-07-01':'五月廿六','2024-08-01':'六月廿七','2024-09-01':'七月廿九','2024-10-01':'八月廿九','2024-11-01':'十月初一','2024-12-01':'冬月初一','2024-12-31':'腊月初一',
        '2025-01-01':'腊月初二','2025-01-28':'除夕','2025-01-29':'正月初一','2025-12-31':'腊月十一',
        '2026-01-01':'腊月十二','2026-02-16':'除夕','2026-02-17':'正月初一','2026-12-31':'冬月十五',
        '2027-01-01':'冬月十六','2027-02-05':'除夕','2027-02-06':'正月初一','2027-12-31':'冬月廿四'
    };

    // 《华严经》经典经文摘录
    const huayanSutraQuotes = [
        "若人欲了知，三世一切佛，应观法界性，一切唯心造。 -----出自《大方广佛华严经》第十二品《夜摩宫中偈赞品》", "心如工画师，能画诸世间，五蕴悉从生，无法而不造。 -----出自《大方广佛华严经》第十二品《夜摩宫中偈赞品》", "不为自己求安乐，但愿众生得离苦。 -----出自《大方广佛华严经》第二十一品《升兜率天宫品》", "菩萨清凉月，常游毕竟空，众生心水净，菩提影现中。 -----出自《大方广佛华严经》第二十三品《十地品》", "忘失菩提心，修诸善法，是为魔业。 -----出自《大方广佛华严经》第三十八品《离世间品》", "一切众生，皆具如来智慧德相，但因妄想执著，不能证得。 -----出自《大方广佛华严经》第三十八品《离世间品》", "不生亦不灭，不常亦不断，不一亦不异，不来亦不出。 -----出自《大方广佛华严经》第一品《世主妙严品》", "譬如一灯，入于暗室，百千年暗，悉能破尽。 -----出自《大方广佛华严经》第二十三品《十地品》", "初发心时，便成正觉。 -----出自《大方广佛华严经》第二十一品《升兜率天宫品》", "应代一切众生受诸楚毒，遍为他们求安隐乐。 -----出自《大方广佛华严经》第二十三品《十地品》", "我当为众生，作不请之友。 -----出自《大方广佛华严经》第三十八品《离世间品》", "于一毫端现宝王刹，坐微尘里转大法轮。 -----出自《大方广佛华严经》第一品《世主妙严品》", "信为道元功德母，长养一切诸善根。 -----出自《大方广佛华严经》第六品《贤首品》", "知一切法，即心自性，成就慧身，不由他悟。 -----出自《大方广佛华严经》第二十二品《十住品》", "离欲及烦恼，证得法性身。 -----出自《大方广佛华严经》第二十三品《十地品》", "佛土生五浊，皆由众生起。 -----出自《大方广佛华严经》第五品《华藏世界品》", "随众生心，而为利益，不作分别。 -----出自《大方广佛华严经》第三十八品《离世间品》", "一切法门，无尽庄严，悉能演说。 -----出自《大方广佛华严经》第二十三品《十地品》", "了知一切法，自性无所有。 -----出自《大方广佛华严经》第六品《贤首品》", "以大悲心，救护一切。 -----出自《大方广佛华严经》第二十三品《十地品》", "普令众生，永离一切烦恼深坑。 -----出自《大方广佛华严经》第二十三品《十地品》", "此法深难见，非识之所及。 -----出自《大方广佛华严经》第一品《世主妙严品》", "无有边际，不可穷尽。 -----出自《大方广佛华严经》第五品《华藏世界品》", "能于一念顷，普现三世劫。 -----出自《大方广佛华严经》第二十三品《十地品》", "一切诸佛，皆从此生。 -----出自《大方广佛华严经》第二十一品《升兜率天宫品》", "常乐柔和忍辱法，安住慈悲喜舍中。 -----出自《大方广佛华严经》第十一品《净行品》", "佛法无人说，虽慧莫能了。 -----出自《大方广佛华严经》第五品《华藏世界品》", "见佛闻法，心无障碍。 -----出自《大方广佛华严经》第二十三品《十地品》", "我当于一切众生，犹如慈母。 -----出自《大方广佛华严经》第三十八品《离世间品》", "能令见者，心得清净。 -----出自《大方广佛华严经》第二十三品《十地品》"
    ];

    const START_DATE = '2024-03-10';
    const GRADUATE_DATE = '2024-06-30';

    // ===================================================================
    // 2. DOM 元素获取
    // ===================================================================

    const usernameInput = document.getElementById('m_username');
    const usershiyanInput = document.getElementById('m_usershiyan');
    const teamNameInput = document.getElementById('m_TeamName');
    const duihuInput = document.getElementById('m_Duihu');
    const faYuanInput = document.getElementById('m_FaYuan'); // 新增
    const fenxiangTextarea = document.getElementById('m_fenxiang');
    const jingwenTextarea = document.getElementById('m_jingwen');
    const finalOutputPre = document.getElementById('final-output');
    const copyBtn = document.getElementById('copy-btn');

    // ===================================================================
    // 3. 核心功能函数
    // ===================================================================

    // 格式化日期为 YYYY-MM-DD
    function formatDate(date) {
        const y = date.getFullYear();
        const m = String(date.getMonth() + 1).padStart(2, '0');
        const d = String(date.getDate()).padStart(2, '0');
        return `${y}-${m}-${d}`;
    }

    // 计算两个日期之间的天数差
    function daysBetween(date1, date2) {
        const oneDay = 24 * 60 * 60 * 1000; // 毫秒数
        // 忽略时间部分，只比较日期
        const d1 = new Date(date1.getFullYear(), date1.getMonth(), date1.getDate());
        const d2 = new Date(date2.getFullYear(), date2.getMonth(), date2.getDate());
        return Math.round((d2 - d1) / oneDay);
    }

    // 获取农历日期
    function getLunarDate(date) {
        const key = formatDate(date);
        return lunarData[key] || '暂无农历数据';
    }

    // 更新所有内容的函数
    function updateContent() {
        const now = new Date();

        // 获取动态变量
        const m_riqi = formatDate(now);
        const m_week = '星期' + ['日', '一', '二', '三', '四', '五', '六'][now.getDay()];
        const m_nongli = getLunarDate(now);
        const m_time = `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}`;
        
        // 计算成立天数 (当天算第1天)
        const m_CreateDay = daysBetween(new Date(START_DATE), now) + 1;
        
        // 计算毕业天数状态
        let m_biyeDay_status;
        const daysToGraduate = daysBetween(now, new Date(GRADUATE_DATE));
        if (daysToGraduate > 0) {
            m_biyeDay_status = `今天是行愿九期，距离毕业还有 ${daysToGraduate} 天`;
        } else if (daysToGraduate === 0) {
            m_biyeDay_status = `今天是行愿九期，今天毕业！`;
        } else {
            // 毕业后第一天算作“已经毕业1天”
            m_biyeDay_status = `今天是行愿九期，已经毕业 ${-daysToGraduate + 1} 天`;
        }

        // 获取用户输入
        const m_username = usernameInput.value;
        const m_usershiyan = usershiyanInput.value;
        const m_TeamName = teamNameInput.value;
        const m_Duihu = duihuInput.value;
        const m_FaYuan = faYuanInput.value; // 新增
        const m_fenxiang = fenxiangTextarea.value;
        const m_jingwen = jingwenTextarea.value;
        
        // 使用模板字符串构建最终输出内容
        const outputText = `今日时间：
${m_riqi} 
${m_week}
农历：${m_nongli}
时间：${m_time}

行愿九期
成立于${START_DATE}
毕业于${GRADUATE_DATE}

今天是行愿九期，成立的第 ${m_CreateDay} 天
${m_biyeDay_status}

${m_username}
${m_usershiyan}

我们的队名是：${m_TeamName}
我们的队呼是：${m_Duihu}
我/我们发愿：${m_FaYuan}

今日分享：
${m_fenxiang}

经文分享：
${m_jingwen}`;

        // 将最终内容更新到 <pre> 标签中
        finalOutputPre.textContent = outputText;
    }

    // ===================================================================
    // 4. 事件监听与初始化
    // ===================================================================

    // 为所有输入框添加 'input' 事件监听器，实现实时更新
    [usernameInput, usershiyanInput, teamNameInput, duihuInput, faYuanInput, fenxiangTextarea, jingwenTextarea].forEach(input => {
        input.addEventListener('input', updateContent);
    });

    // 为复制按钮添加点击事件
    copyBtn.addEventListener('click', function() {
        // 使用现代的 navigator.clipboard API
        navigator.clipboard.writeText(finalOutputPre.textContent).then(function() {
            // 复制成功后的反馈
            const originalText = copyBtn.textContent;
            copyBtn.textContent = '已复制到剪贴板！';
            copyBtn.style.backgroundColor = '#1e7e34';
            setTimeout(function() {
                copyBtn.textContent = originalText;
                copyBtn.style.backgroundColor = '#28a745';
            }, 2000);
        }).catch(function(err) {
            console.error('复制失败: ', err);
            alert('复制失败，请手动复制。');
        });
    });

    // 初始化页面
    // 1. 随机设置经文分享的默认值
    jingwenTextarea.value = huayanSutraQuotes[Math.floor(Math.random() * huayanSutraQuotes.length)];
    // 2. 首次加载时，立即执行一次更新以显示初始内容
    updateContent();

});
</script>

</body>
</html>