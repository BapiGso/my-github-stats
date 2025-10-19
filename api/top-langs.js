import { extractSvgContent, getDecorationSvg, fetchGithubStats } from './utils.js';

export default async function handler(req, res) {
    try {
        const {
            username,
            decoration = 'cat',
            decoration_x,
            decoration_y,
            svg_width = '500',
            svg_height = '170',
            ...otherParams
        } = req.query;

        if (!username) {
            return res.status(400).json({ error: 'Username is required' });
        }

        // 构建参数
        const langsParams = {
            username,
            layout: otherParams.layout || 'compact',
            card_width: otherParams.card_width || '350',
            hide_border: otherParams.hide_border || 'true',
            show_bg: otherParams.show_bg || '1',
            ...otherParams
        };

        // 获取原始 SVG
        const originalSvg = await fetchGithubStats(
            'https://github-readme-stats.vercel.app/api/top-langs/',
            langsParams
        );

        // 获取装饰
        const decorationSvg = decoration === 'none'
            ? ''
            : getDecorationSvg(decoration, decoration_x, decoration_y);

        // 创建增强版 SVG
        const enhancedSvg = `
<svg width="${svg_width}" height="${svg_height}" xmlns="http://www.w3.org/2000/svg">
  <rect
    x="0.5"
    y="0.5"
    rx="4.5"
    height="99%"
    width="${parseInt(svg_width) - 1}"
    fill="#fffefe"
    stroke="#e4e2e2"
    stroke-opacity="1"
  />
  <image href="https://github-readme-stats.vercel.app/api/top-langs/?${new URLSearchParams(langsParams).toString()}"/>
  ${decorationSvg}
</svg>`.trim();

        // 返回 SVG
        res.setHeader('Content-Type', 'image/svg+xml');
        res.setHeader('Cache-Control', 's-maxage=3600, stale-while-revalidate');
        res.status(200).send(enhancedSvg);

    } catch (error) {
        console.error('Error:', error);
        res.status(500).json({ error: error.message || 'Internal server error' });
    }
}