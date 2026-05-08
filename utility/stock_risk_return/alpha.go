package stock_risk_return

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
)

// Alpha 计算詹森阿尔法（Jensen's Alpha）
// 公式：α = Rp - [Rf + β * (Rm - Rf)]
// Rp: 组合实际收益率 (年)
// Rf: 无风险利率 (年)
// Rm: 市场基准收益率 (年)
// Beta: 组合的贝塔系数
// 詹森阿尔法 = 实际收益 - 预期收益
// 预期收益 = Rf + Beta * (Rm - Rf)
func Alpha(ctx context.Context, Rp, Rf, Rm, Beta float64) float64 {
	g.Log().Infof(ctx, "Rp: %f,  Rf: %f, Rm: %f, Beta: %f", Rp, Rf, Rm, Beta)
	expectedReturn := Rf + Beta*(Rm-Rf)
	alpha := Rp - expectedReturn
	// 保留 4 位小数
	return float64(int(alpha*1000000)) / 1000000
}
