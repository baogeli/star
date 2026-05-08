import requests
import json
from datetime import datetime

def get_stock_quote(symbol):
    url = "https://stock.xueqiu.com/v5/stock/quote.json"
    params = {
        "symbol": symbol,
        "extend": "detail"
    }

    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
        "Accept": "*/*",
        "Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
        "Origin": "https://xueqiu.com",
        "Referer": f"https://xueqiu.com/S/{symbol}",
    }

    # Cookies converted from the curl command string
    cookies = {
        "device_id": "729d5181bebe9efb9020d894e4697613",
        "xq_a_token": "ca35d6d2fa5e735759056fc62797546c18062187",
        "u": "961767857278498",
    }

    try:
        response = requests.get(url, headers=headers, params=params, cookies=cookies)
        
        if response.status_code != 200:
            print(f"请求失败，状态码：{response.status_code}")
            return None
            
        data = response.json()
        
        # Parse and return specific quote data
        if data.get('data') and data['data'].get('quote'):
            return data['data']['quote']
        
        return data
        
    except requests.exceptions.RequestException as e:
        print(f"Error fetching URL: {e}")
        return None

if __name__ == "__main__":
    symbol = "SH508033"
    print(f"正在获取 {symbol} 的数据...")
    quote = get_stock_quote(symbol)
    
    if quote:
        # Display key information
        print("-" * 30)
        print(f"股票名称: {quote.get('name')}")
        print(f"股票代码: {quote.get('symbol')}")
        print(f"当前价格: {quote.get('current')}")
        print(f"涨跌幅:   {quote.get('percent')}%")
        print(f"涨跌额:   {quote.get('chg')}")
        print(f"更新时间: {datetime.fromtimestamp(quote.get('timestamp', 0)/1000).strftime('%Y-%m-%d %H:%M:%S')}")
        print("-" * 30)
