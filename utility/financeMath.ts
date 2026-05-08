/**
 * 计算年化收益率 (基于自然日)
 * 公式: (1 + 期间总收益率)^(365 / 自然日数) - 1
 * @param totalReturn 期间总收益率
 * @param durationDays 自然日数
 */
export const calculateAnnualizedReturn = (totalReturn: number, durationDays: number): number => {
    if (durationDays <= 0) return 0;
    return Math.pow(1 + totalReturn, 365 / durationDays) - 1;
};

/**
 * 计算年化波动率
 * 公式: 日收益率标准差 * sqrt(250)
 * @param dailyReturns 日收益率数组
 */
export const calculateVolatility = (dailyReturns: number[]): number => {
    if (dailyReturns.length < 2) return 0;
    const mean = dailyReturns.reduce((a, b) => a + b, 0) / dailyReturns.length;
    const variance = dailyReturns.reduce((a, b) => a + Math.pow(b - mean, 2), 0) / (dailyReturns.length - 1);
    return Math.sqrt(variance) * Math.sqrt(250);
};

/**
 * 计算最大回撤及相关周期
 * Formula: Max Drawdown = min(Current Value - Peak Value) / Peak Value
 * @param netValues 净值数组
 */
export const calculateMaxDrawdown = (netValues: number[]): { maxDrawdown: number, formationPeriod: number, recoveryPeriod: number } => {
    let maxDd = 0;
    let peak = -Infinity;
    let peakIndex = 0;
    let maxDdEndIndex = 0;

    netValues.forEach((val, i) => {
        if (val > peak) {
            peak = val;
            peakIndex = i;
        }
        const dd = (val - peak) / peak;
        if (dd < maxDd) {
            maxDd = dd;
            maxDdEndIndex = i;
        }
    });

    // 回撤形成期: 也就是从最高点到最大回撤点的时间
    const formationPeriod = maxDdEndIndex - peakIndex;
    // 回撤修复期: 从最大回撤点到当前的时间 (如果已修复则为修复耗时，此处简化计算)
    const recoveryPeriod = netValues.length - 1 - maxDdEndIndex;

    return { maxDrawdown: maxDd, formationPeriod, recoveryPeriod };
};

/**
 * 计算贝塔系数 (Beta)
 * 衡量组合相对于市场的敏感度
 * 公式: Cov(Rp, Rm) / Var(Rm)
 * @param returns 组合收益率数组
 * @param benchmarkReturns 基准收益率数组
 */
export const calculateBeta = (returns: number[], benchmarkReturns: number[]): number => {
    if (returns.length !== benchmarkReturns.length || returns.length < 2) {
        console.log("⚠️ 数组长度不一致或不足 2，返回 0");
        return 0;
    }

    // 取最小长度，防止数组长度不一致
    const length = Math.min(returns.length, benchmarkReturns.length);
    const meanRp = returns.reduce((a, b) => a + b, 0) / length;
    const meanRm = benchmarkReturns.reduce((a, b) => a + b, 0) / length;

    let cov = 0;
    let varRm = 0;

    for (let i = 0; i < length; i++) {
        cov += (returns[i] - meanRp) * (benchmarkReturns[i] - meanRm);
        varRm += Math.pow(benchmarkReturns[i] - meanRm, 2);
    }

    return varRm < 1e-8 ? 0 : cov / varRm;
};

/**
 * 计算下行风险 (Downside Risk)
 * 仅考虑负收益的标准差
 * Formula: sqrt(sum(min(0, Rp - Rf)^2) / N) * sqrt(250)
 * @param returns 收益率数组
 * @param rf 无风险利率
 */
export const calculateDownsideRisk = (returns: number[], rf: number = 0): number => {
    const negativeReturns = returns.filter(r => r < rf);
    if (negativeReturns.length === 0) return 0;

    const sumSqDiff = negativeReturns.reduce((acc, r) => acc + Math.pow(r - rf, 2), 0);
    // 通常这里除以 N 或 N-1
    return Math.sqrt(sumSqDiff / returns.length) * Math.sqrt(250);
};

/**
 * 综合计算所有风险收益指标
 * @param portfolioValues 组合每日净值数组
 * @param benchmarkValues 基准每日净值数组
 * @param dates 日期数组
 * @param rf 无风险利率 (默认 2%)
 */
export const calculateMetrics = (
    portfolioValues: number[],
    benchmarkValues: number[],
    dates: string[], // 新增日期数组，用于计算自然日
    rf: number = 0.02 // 无风险利率 2%
) => {
    console.log("\n=== calculateMetrics 输入数据 ===");
    console.log("portfolioValues 长度:", portfolioValues.length);
    console.log("benchmarkValues 长度:", benchmarkValues.length);
    console.log("dates 长度:", dates.length);
    
    if (portfolioValues.length < 2) {
        console.log("⚠️ portfolioValues 长度不足 2，返回 null");
        return null;
    }

    // 计算日收益率
    const pReturns: number[] = [];
    const bReturns: number[] = [];
    for (let i = 1; i < portfolioValues.length; i++) {
        pReturns.push((portfolioValues[i] - portfolioValues[i - 1]) / portfolioValues[i - 1]);
        if (benchmarkValues[i] && benchmarkValues[i - 1]) {
            bReturns.push((benchmarkValues[i] - benchmarkValues[i - 1]) / benchmarkValues[i - 1]);
        } else {
            bReturns.push(0);
        }
    }

    // 计算自然日天数
    const startDate = new Date(dates[0]);
    const endDate = new Date(dates[dates.length - 1]);
    // 毫秒转天数
    const durationDays = Math.max(1, (endDate.getTime() - startDate.getTime()) / (1000 * 60 * 60 * 24));

    // 年化收益率 (Annualized Return) - 使用自然日 365
    // 累计收益率 (Cumulative Return)
    const cumulativeReturn = (portfolioValues[portfolioValues.length - 1] - portfolioValues[0]) / portfolioValues[0];
    const annualizedReturn = calculateAnnualizedReturn(cumulativeReturn, durationDays);

    // 年化波动率 (Volatility)
    const volatility = calculateVolatility(pReturns);
    // 贝塔系数 (Beta)
    const beta = calculateBeta(pReturns, bReturns);

    // 最大回撤 (Max Drawdown)
    const { maxDrawdown, formationPeriod, recoveryPeriod } = calculateMaxDrawdown(portfolioValues);

    // 詹森阿尔法 (Jensen's Alpha)
    // 公式：α = Rp - [Rf + β * (Rm - Rf)]
    // 使用年化值
    // 基准年化收益
    const benchmarkCumulativeReturn = (benchmarkValues[benchmarkValues.length - 1] - benchmarkValues[0]) / benchmarkValues[0];
    const benchmarkAnnualizedReturn = calculateAnnualizedReturn(benchmarkCumulativeReturn, durationDays);
    const jensenAlpha = annualizedReturn - (rf + beta * (benchmarkAnnualizedReturn - rf));
    
    console.log("=== 基准收益计算 ===");
    console.log("基准累计收益率:", benchmarkCumulativeReturn);
    console.log("基准年化收益率:", benchmarkAnnualizedReturn);
    console.log("组合年化收益率:", annualizedReturn);
    console.log("阿尔法 (超额收益):", jensenAlpha);
    console.log("===================");

    // 夏普比率 (Sharpe Ratio)
    // 公式: (Rp - Rf) / σp
    const sharpeRatio = volatility < 1e-8 ? 0 : (annualizedReturn - rf) / volatility;

    // 跟踪误差 (Tracking Error)
    // 公式: Stdev(Rp - Rm)
    const diffReturns = pReturns.map((r, i) => r - bReturns[i]);
    const trackingError = calculateVolatility(diffReturns); // 辅助函数已年化

    // 信息比率 (Information Ratio)
    // 公式: (Rp - Rm) / Tracking Error
    const infoRatio = trackingError < 1e-8 ? 0 : (annualizedReturn - benchmarkAnnualizedReturn) / trackingError;

    // 索提诺比率 (Sortino Ratio)
    // 公式: (Rp - Rf) / 下行标准差
    const downsideRisk = calculateDownsideRisk(pReturns, rf / 250); // 转为日频无风险利率
    const sortinoRatio = downsideRisk < 1e-8 ? 0 : (annualizedReturn - rf) / downsideRisk;

    // 特雷诺比率 (Treynor Ratio)
    // 公式: (Rp - Rf) / β
    const treynorRatio = Math.abs(beta) < 1e-8 ? 0 : (annualizedReturn - rf) / beta;

    // R平方 (R-Squared)
    const rSquared = calculateRSquared(pReturns, bReturns);

    // 胜率 (Win Rate)
    const positiveDays = pReturns.filter(r => r > 0).length;
    const winRate = positiveDays / pReturns.length;

    // 在险价值 (VaR)
    // 历史模拟法 95% 置信度
    const sortedReturns = [...pReturns].sort((a, b) => a - b);
    const index95 = Math.floor(sortedReturns.length * 0.05);
    const var95 = sortedReturns[index95]; // 日度 VaR


    // 辅助函数：安全格式化比率
    const safeFormat = (val: number, decimals: number = 2): string => {
        if (!Number.isFinite(val) || Math.abs(val) > 10000) return (0).toFixed(decimals);
        return val.toFixed(decimals);
    };

    return {
        riskReturn: {
            alpha: Number(safeFormat(jensenAlpha, 4)),
            cumulativeReturn,
            annualizedReturn,
            benchmarkAnnualizedReturn,  // ⭐ 新增：基准年化收益率
            sharpeRatio: safeFormat(sharpeRatio),
            infoRatio: safeFormat(infoRatio),
            sortinoRatio: safeFormat(sortinoRatio),
            jensenAlpha: safeFormat(jensenAlpha, 4),
            treynorRatio: safeFormat(treynorRatio),
            winRate,
            positivePeriods: positiveDays,
            totalPeriods: pReturns.length
        },
        risk: {
            beta: safeFormat(beta),
            annualizedVol: volatility,
            trackingError,
            downsideRisk,
            var: Math.abs(var95),
            maxDrawdown,
            drawdownFormation: formationPeriod,
            drawdownRecovery: String(recoveryPeriod),
            consecutiveDrop: 0, // 暂时保留占位，后续可添加计算逻辑
            rSquare: safeFormat(rSquared)
        }
    };
};

/**
 * 计算 R平方 (R-Squared)
 * 衡量组合收益率与基准收益率的相关性
 * 公式: Correlation(Rp, Rm)^2
 * @param returns 组合收益率数组
 * @param benchmarkReturns 基准收益率数组
 */
export const calculateRSquared = (returns: number[], benchmarkReturns: number[]): number => {
    if (returns.length !== benchmarkReturns.length || returns.length < 2) return 0;

    const meanRp = returns.reduce((a, b) => a + b, 0) / returns.length;
    const meanRm = benchmarkReturns.reduce((a, b) => a + b, 0) / benchmarkReturns.length;

    let cov = 0;
    let varRp = 0;
    let varRm = 0;

    for (let i = 0; i < returns.length; i++) {
        const diffRp = returns[i] - meanRp;
        const diffRm = benchmarkReturns[i] - meanRm;
        cov += diffRp * diffRm;
        varRp += Math.pow(diffRp, 2);
        varRm += Math.pow(diffRm, 2);
    }

    if (varRp === 0 || varRm === 0) return 0;

    // 相关系数 Correlation = Cov / (StdRp * StdRm)
    // R2 = Correlation^2 = Cov^2 / (VarRp * VarRm)
    return Math.pow(cov, 2) / (varRp * varRm);
};
