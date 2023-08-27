import dotenv from "dotenv"
import express from "express"
import cors from "cors"
import { init as ldapInit, query as ldapQuery } from "./ldap"
import { version } from "./package.json"
import auth from "@moreillon/express_identification_middleware"

dotenv.config()

ldapInit()

const { PORT = 80, IDENTIFICATION_URL } = process.env

const app = express()
app.use(express.json())
app.use(cors())

app.get("/", async (req, res) => {
  res.send({
    version,
    auth: {
      identification_url: IDENTIFICATION_URL,
    },
  })
})

if (IDENTIFICATION_URL) {
  const authOptions = { url: IDENTIFICATION_URL }
  app.use(auth(authOptions))
}

app.get("/search", async (req, res) => {
  const { query }: any = req

  const filter = Object.keys(query)
    .map((key) => `(${key}=${query[key]})`)
    .join()

  const options = {
    filter: `(&${filter})`,
    scope: "sub",
    paged: true,
    sizeLimit: 1000,
  }

  const records = await ldapQuery(options)

  res.send(records)
})

app.listen(PORT, () => {
  console.log("[Express] Listening on port ${PORT}")
})
