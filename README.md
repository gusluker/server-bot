# SERVER BOT

Servicio de organización y redireccionamiento de información. El servicio ServerBot organiza y redirecciona la información acorde a su localización geográfica y los correos enlazados a esa localización en el servicio.

## Configuración

La configuración se realiza en un archivo elegido por el usuario con formato `.sbot`, con ruta absoluta especificada en la variable de entorno **SBOT_CONFIG_FILE_PATH**. La definición de la variable es obligatoria.

### Nivel de registro

**log-level**. Nivel `Info` por defecto. Indica el nivel mínimo que se registrará en la bitácora. Si elige `Debug`, registrará este nivel y todos los niveles que estén por debajo de el, los cuales son: `Info`, `Warn`, `Error`, `Fatal` y `Panic`. No se diferencia entre mayúsculas y minúsculas:

  * Trace
  * Debug
  * Info 
  * Warn
  * Error
  * Fatal
  * Panic

```json
{
	"log-level":"Debug"
}
```

### Ruta de bitácora

**log-path** Syslog por defecto en Linux, obligatoria para Windows. Indica la ruta absoluta de un archivo de texto, para el registro de la bitácora del servicio, si el archivo no existe, será creado por el sistema. Asegúrese de tener permisos de escritura y lectura en la ruta seleccionada. 

### Lista de correos

**mail-list** Obligatoria. Lista de correos a la que se enviará la información previamente organizada. Solo el correo indexado a **world** es obligatorio. Para obtener una ruta de geolocalización, la cual posee la forma "código.estado.ciudad", utilice el script `script/get` al cual debe pasarle como parámetros la latitud y longitud geográfica, en ese órden. Estos valores los puede obtener de Google Maps, por ejemplo.

La opción **mail-list** posee la siguiente estructura:

```json
{
	"mail-list": {
		"world": "correo0@correo.com",
		"ec.Pichincha.Quito":"correo1@proton.com",
		"ec.Manabí.Manta":[
			"correo2@proton.com",
			"correo3@proton.com"
		],
		"ec.Manabí.Jaramijó":"correo4@proton.com",
		"ec.Manabí":[
			"correo5@proton.com",
			"correo6@proton.com"
		],
		"ec":"correo7@proton.com"
	}
}
```

Internamente se creará un arbol donde la primera jerarquía será **world**, la segunda jerarquía pertenece a los paises, la tercera a los estados y la última a las ciudades. Es posible indexar correos a cualquier jerarquía de este árbol. Los datos ingresados en sistema, buscarán una igualdad en la jerarquía, partiendo por las ciudades, en caso de no encontrarla, subirán a los estados, países, hasta llegar a **world**. Por eso esta última es obligatoria.

### Configuración de Correos de envío

Obligatoria. Actualmente solo se usa correos Gmail mediante OAuth para realizar la conexión con el cliente gmail y envíar los correos eléctronicos. Se pueden agregar varios correos en caso que la conexión sea bloqueada por superar el límite de mensajes diarios.

La configuración se realiza en un archivo `.sbot` el cual su ruta debe agregarse a la variable del sistema **SBOT_CONFIG_FILE_EMAILS**.

El archivo posee la siguiente estrucutra JSON:

```json
[
	{
		"Email": "correo1@correo.com",
		"ClientID": "IdClient1",
		"ClientSecret": "ClientSecret1",
		"AccessToken": "Access1",
		"RefreshToken": "Token1",
		"TokenType": "Type"
	},
	{
		"Email": "correo2@correo.com",
		"ClientID": "IdClient2",
		"ClientSecret": "ClientSecret2",
		"AccessToken": "Access2",
		"RefreshToken": "Token2",
		"TokenType": "Type"
	}
]
```

## Documentación

La inicialización del servicio ServerBot se realizar mediante la función `New([]plugins.Plugin, *gin.Engine) *Server`, este inicializa el servicio y lo enruta al recurso `/sorter`. Retorna un **panic** si encuentra un error en la configuración.

El recurso `/sorter` **solo** recibe peticiones POST con contenido de tipo `application/json` que posean la siguiente estructura:

```json
{
	"index":"identificador de la petición",
	"plug":"nombre plugin",
	"coord":{
		"lat":"número latitud geográfica",
		"lon":"número longitud geográfica",
	},
	"location":{
		"coord": {
			"lat":"número latitud geográfica",
			"lon":"número longitud geográfica"
		},
		"country":"Pais",
		"country_code":"Código",
		"state":"Estado",
		"city":"Ciudad"
	},
	"data":{ ...cualquier estructura json válida... }
}
```

- **index**. Nombre que se usará para registrar los datos en sistemas.
- **plug**. Nombre del plugin que procesará los datos antes de ser enviados por SMTP.
- **coord**. Puede no ser definida si se define **location**. Coordenadas geográfica que se usarán para consultar la información geográfica en el servicio de georeferenciación.
- **location**. Puede no ser definida si se define **coord**. Este campo se define para que el servidor no realice una consulta al servicio de georeferenciación, el cual es muy lento. Use como ejemplo el script **script/get** para saber cómo consultar esta información.
- **data**. Datos cualquiera que serán procesados por el plugin que concuerde con el nombre en **plug**.

### Plugins

El recurso `sorter` recibe los datos, los organiza y reenvía según sus coordenadas geográfica. Para poder realizar el envío de datos, se debe procesar el parámetro `data` el cual es un dato personalizado enviado por el usuario que posee valores que solo él conoce. Para eso, se usan los plugins.

Los plugins son pasados como parámetros al inicializar el servicio y son usados para procesar los datos recibidos y organizados por el recurso `sorter`. Los plugins son tipos de datos que heredan la interfaz `Plug`, la cual posee la siguiente estructura:

```golang
type Plug interface {
	IsThisPlugin(data *models.Data)	bool
	Run(data *models.Data)		([]*models.GData, error)
	GetName() 			string
}

//Tipo de dato en server/models/data.go
type Data struct {
	Index		string 		
	NamePlug	string 	
	Coord		*GeoCoord	
	Loc		*Location
	Body		interface{}	
}

//Tipo de dato en server/models/location.go
type Location struct {
	Coord 		*GeoCoord
	Country 	string
	CountryCode	string
	State		string
	City		string
}

//Tipo de dato en server/models/geocoord.go
type GeoCoord struct {
	Latitude	string
	Longitude 	string
}
```
- **IsThisPlug**. Función que recibe un tipo de dato `Data` y retorna un booleano `true` en caso que pueda ser procesado por este plugin.
- **Run**. Función que recibe un tipo de dato `Data` y retorna un array de `GData` listos para su envío a través de smtp. En caso contrario, retorna error.
- **GetName**. Función que retorna el nombre del plugin. Este nombre debe concordar con el nombre recibido en el parámetro **plug**.

La función **Run** procesa los datos y retorna un tipo de dato `GData` usado para el envío a través de smtp. Este posee la siguiente estructura:

```golang
type GData struct {
	ContentType 		string
	ContentTransferEncoding	string
	Name			string
	Data 			string
}
```

Un `GData` representa un archivo adjunto o texto, el cual será indicado por sus respectivo valores:

- **ContentType**. Encabezado que representa qué [tipo de medio](https://es.wikipedia.org/wiki/Multipurpose_Internet_Mail_Extensions#Content-Type) representa el mensaje.
- **ContentTransferEncoding**. Encabezado que representa el [tipo de condificación](https://es.wikipedia.org/wiki/Multipurpose_Internet_Mail_Extensions#Content-Transfer-Encoding) usada en el mensaje.
- **Name**. Encabezado que representa el nombre del mensaje. Este campo se puede dejar vaciío para `ContentType: text/plain`.
- **Data**. Encabezado que representa el archivo adjunto. Su condificación es representada por `ContentTransferEncoding`.
