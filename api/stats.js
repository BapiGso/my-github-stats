export default async function handler(req, res) {
    try {
        // 获取查询参数
        const { username, ...otherParams } = req.query;

        if (!username) {
            return res.status(400).json({ error: 'Username is required' });
        }

        // 构建 GitHub Stats API URL
        const params = new URLSearchParams({
            username,
            show_icons: otherParams.show_icons || 'true',
            hide_rank: otherParams.hide_rank || 'true',
            hide_border: otherParams.hide_border || 'true',
            card_width: otherParams.card_width || '350',
            ...otherParams
        });

        const githubStatsUrl = `https://github-readme-stats.vercel.app/api?${params.toString()}`;

        // 获取原始 SVG
        const response = await fetch(githubStatsUrl);
        if (!response.ok) {
            throw new Error('Failed to fetch GitHub stats');
        }

        const originalSvg = await response.text();

        // 创建带有自定义装饰的新 SVG
        const enhancedSvg = `
<svg width="500" height="200" xmlns="http://www.w3.org/2000/svg">
  <rect
    x="0.5"
    y="0.5"
    rx="4.5"
    height="98%"
    width="499"
    fill="#fffefe"
    stroke="#e4e2e2"
    stroke-opacity="1"
  />
  
  <!-- 嵌入原始 GitHub Stats -->
  <g>
    ${extractSvgContent(originalSvg)}
  </g>
  
  <!-- 自定义装饰路径 -->
  <svg x="400" y="0" width="80" viewBox="0 0 230.7 357.6" fill-opacity="0.2">
    <path d="M1045.4,483.4c40.5-5.7,81.4-2,122.1-3.7,7.4-.1,7.4-.1,7.5-7.6s4.5-14.6,12.6-12.1-.5,17.9,8.6,19.2c12.6.7,25.2,2.6,37.6-2,8.9-3.3,12.9-.5,13.3,9s-1.8,15.3-11.5,11.7c-11.5-3.7-23.1-2.9-34.7-2.6-8.3.3-8.4.6-8.9,8.9s-3.2,11-11.5,9c-11.2-4.4,1.7-18.9-16.5-17.1C1124.9,493.3,1080.5,504.6,1045.4,483.4Z"
      transform="translate(-1045.4 -302.8)">
      <animateTransform
        attributeName="transform"
        attributeType="XML"
        type="translate"
        values="-1045.4,-302.8; -1045.4,-320.8; -1045.4,-310.8; -1045.4,-325.8; -1045.4,-302.8"
        keyTimes="0; 0.25; 0.5; 0.75; 1"
        dur="5s"
        repeatCount="indefinite"/>
    </path>
  </svg>
</svg>`.trim();

        // 返回 SVG
        res.setHeader('Content-Type', 'image/svg+xml');
        res.setHeader('Cache-Control', 's-maxage=3600, stale-while-revalidate');
        res.status(200).send(enhancedSvg);

    } catch (error) {
        console.error('Error:', error);
        res.status(500).json({ error: 'Internal server error' });
    }
}

// 提取 SVG 内容（去除外层 svg 标签）
function extractSvgContent(svgString) {
    const match = svgString.match(/<svg[^>]*>([\s\S]*)<\/svg>/i);
    return match ? match[1] : svgString;
}