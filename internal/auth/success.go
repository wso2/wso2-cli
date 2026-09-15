package auth

const oauthSuccessPage = `
<!doctype html>
<meta charset="utf-8">
<title>Success: Integration Platform CLI</title>
<style type="text/css">
body {
  color: #1B1F23;
  background: #F6F8FA;
  font-size: 14px;
  font-family: -apple-system, "Segoe UI", Helvetica, Arial, sans-serif;
  line-height: 1.5;
  max-width: 620px;
  margin: 28px auto;
  text-align: center;
}

h1 {
  font-size: 24px;
  margin-bottom: 0;
}

p {
  margin-top: 0;
}

.brand {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14px;
}

.brand-name {
  text-align: left;
  line-height: 1.1;
}

.brand-org {
  display: block;
  font-size: 30px;
  letter-spacing: 0.5px;
}

.brand-product {
  display: block;
  font-size: 22px;
  font-weight: 600;
}

.box {
  border: 1px solid #E1E4E8;
  background: white;
  padding: 24px;
  margin: 28px;
}
</style>
<body>
  <div class="brand">
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="20 20 75 75" width="56" height="56"
         role="img" aria-label="WSO2">
      <path d="M62.54,72.25s-.07,0-.1,0c-.71-.04-1.31-.53-1.5-1.21l-5.34-19.59-2.74,9.14c-.21.7-.85,1.18-1.58,1.18h-10.25c-.91,0-1.65-.74-1.65-1.65s.74-1.65,1.65-1.65h9.02l4.04-13.49c.21-.7.85-1.18,1.58-1.18,0,0,.01,0,.02,0,.74,0,1.38.51,1.57,1.22l5.58,20.45,2.42-5.96c.25-.62.86-1.03,1.53-1.03h8.21c.91,0,1.65.74,1.65,1.65s-.74,1.65-1.65,1.65h-7.1l-3.83,9.46c-.25.63-.86,1.03-1.53,1.03Z"/>
        <path d="M58.01,87.05c-16.01,0-29.03-13.02-29.03-29.03s13.02-29.03,29.03-29.03,29.04,13.03,29.04,29.03-13.03,29.03-29.04,29.03ZM58.01,32.28c-14.19,0-25.73,11.54-25.73,25.73s11.54,25.73,25.73,25.73,25.73-11.54,25.73-25.73-11.54-25.73-25.73-25.73Z"/>
    </svg>
    <div class="brand-name">
      <span class="brand-org">WSO2</span>
      <span class="brand-product">Integration Platform</span>
    </div>
  </div>
  <div class="box">
    <h1>Successfully authenticated Integration Platform CLI</h1>
    <p>You may now close this tab and return to the terminal.</p>
  </div>
</body>
`
