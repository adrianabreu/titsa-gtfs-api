import { logger, schedules, tasks } from "@trigger.dev/sdk/v3";
import { XMLParser } from "fast-xml-parser";

interface Llegada {
  codigoParada: string;
  denominacion: string;
  destinoLinea: string;
  hora: string;
  idTrayecto: string;
  linea: string;
  minutosParaLlegar: string;
}

interface LlegadasResponse {
  llegadas: {
    llegada: Llegada | Llegada[];
  };
}

interface ParsedArrival {
  line: string;
  stop_id: number;
  estimated_arrival: Date;
}

function parseArrivalIntoJson(arrival: Llegada): ParsedArrival {
  // Parse hora format: "03/11/2025 09:05:30" (dd/mm/yyyy HH:mm:ss)
  const [datePart, timePart] = arrival.hora.split(" ");
  const [day, month, year] = datePart.split("/");
  const [hours, minutes, seconds] = timePart.split(":");
  
  // Create Date in GMT timezone (Canary Islands)
  // Note: JS Date uses local timezone, but will be serialized to UTC for Tinybird
  const arrivalTime = new Date(Date.UTC(+year, +month - 1, +day, +hours, +minutes, +seconds));
  arrivalTime.setMinutes(arrivalTime.getMinutes() + +arrival.minutosParaLlegar);
  
  return {
    line: arrival.linea,
    stop_id: +arrival.codigoParada,
    estimated_arrival: arrivalTime,
  };
}

async function publishToTinybird(
  token: string,
  data: ParsedArrival[]
): Promise<void> {
  const TINYBIRD_URL = "https://api.tinybird.co/v0/events?name=realtime_arrivals";
  
  // Convert to NDJSON format with explicit serialization
  const ndjson = data.map(item => {
    // Explicitly serialize to ensure proper types for Tinybird
    const record = {
      line: String(item.line), // Force string type
      stop_id: item.stop_id,
      estimated_arrival: item.estimated_arrival.toISOString(), // Convert Date to ISO string
    };
    return JSON.stringify(record);
  }).join('\n');
  
  const response = await fetch(TINYBIRD_URL, {
    method: "POST",
    headers: {
      "Authorization": `Bearer ${token}`,
    },
    body: ndjson,
  });
  
  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(
      `Tinybird API error: ${response.status} - ${errorText}`
    );
  }
  
  logger.log("Successfully published to Tinybird", { count: data.length });
}

export const fetchTitsaArrivals = schedules.task({
  id: "fetch-titsa-arrivals",
  cron: "*/15 * * * *",
  run: async (payload: any, { ctx }) => {
    const idApp = process.env.TITSA_ID_APP;
    
    if (!idApp) {
      throw new Error("TITSA_ID_APP environment variable is not set");
    }

    const stop = 1918

    const url = `https://apps.titsa.com/apps/apps_sae_llegadas_parada.asp?idApp=${idApp}&idParada=${stop}`;
    
    logger.log("Fetching TITSA arrivals", { idParada: payload.idParada, url });

    try {
      const response = await fetch(url);
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const xmlText = await response.text();

      const parser = new XMLParser({
        ignoreAttributes: false,
        attributeNamePrefix: "@_",
        textNodeName: "#text",
      });

      const jsonObj: LlegadasResponse = parser.parse(xmlText);
      
      const llegadas = Array.isArray(jsonObj.llegadas.llegada)
        ? jsonObj.llegadas.llegada
        : jsonObj.llegadas.llegada
        ? [jsonObj.llegadas.llegada]
        : [];

      // Parse arrivals into the desired format with calculated timestamps
      const parsedArrivals = llegadas.map(parseArrivalIntoJson);
      
      logger.log("Parsed arrivals with timestamps", {
        count: parsedArrivals.length,
        arrivals: parsedArrivals,
      });

      // Optionally publish to Tinybird
      const tinybirdToken = process.env.TINYBIRD_TOKEN;
      if (!tinybirdToken) {
        throw new Error(
          "TINYBIRD_TOKEN environment variable is required when publishToTinybird is true"
        );
      }
      await publishToTinybird(tinybirdToken, parsedArrivals);

    } catch (error) {
      logger.error("Error fetching TITSA arrivals", { error });
      throw error;
    }
  },
});
