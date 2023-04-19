# GEX

GEX is an open source crypto-currency exchange Software

## What is GEX

Gex is an open-source crypto currency exchange software implementation in Go. 
This project is inspired mainly by  (Peatio)[https://github.com/peatio/peatio]  
Just like Peatio this software is designed to work with external frontend and 
server components. 

## Features

- High performance matching and trading engine
- Accounting and order management
- Multiple wallet support [WIP] 
- Coin API [WIP] 
- Kafka event API   
- Usability and scalability 
- Websocket API and high frequency trading support [WIP] 
- Support for multiple digital curreneies (Bitcoin, ETH, Cardano, etc..) 
- Support for ERC20 tokens 
- API endpoint for FIAT deposits and payment gateways 
- Admin dashboard and management tools cfr [Gex-Web](https://github.com/irononet/gex-admin) 
- Top notch security
- Well-documented [WIP]

## System Overview 


## Getting started


### Minimalistic local development environment 

#### Prerequisites 

* [Docker](https://docs.docker.com/install) installed 
* [Docker compose](https://docs.docker.com/compose/install/) installed 
* Go 1.19 

### Install 

To get started with go-exchange,  
* git clone https://github.com/irononet/go-exchange.git 
* RUN cd go-exchange
* RUN go mod install // to install dependencies 
* RUN go run main.go 

To run the exchange with Docker-Compose
* RUN cd go-exchange 
* RUN docker-compose up 

## Documentation

The go-exchange documentation is still a work in progress at this point. You can view the documentation on the [official go-exchange website](https://irononet.github.io/go-exchange/). [WIP]

## Contributing

If you'd like to contribute to go-exchange, please submit a pull request with your changes or open an issue to discuss potential changes. We welcome contributions of all kinds, including bug fixes, new features, and documentation improvements.

## License

GEX is released under the MIT license. See [LICENSE](https://github.com/irononet/go-exchange/blob/master/LICENSE) for more information.

## Contact

If you have any questions or feedback about go-exchange, please feel free to contact us at arnaudwanetwork9@gmail.com We'd love to hear from you!

## Acknowledgments

- This project was inspired by the excellent:
  - [gitbitex-spot](https://github.com/gitbitex/gitbitex-spot) by Gitbitex
  - [peatio](https://github.com/openware/peatio) by Openware 

- Thanks to the Go community for providing such a great language and ecosystem for building high-quality software.
