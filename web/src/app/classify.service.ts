import {Injectable} from '@angular/core';
import {Http, Response, RequestOptions, Headers} from '@angular/http';
import {Observable} from 'rxjs/Rx';
import {Collection} from './collections/collection';

export enum WebSocketStatus {
    NONE = 0,
    CONNECTING,
    OPEN,
    CLOSING,
    CLOSE,
    ERROR,
}

@Injectable()
export class ClassifyService {

    private url = "http://localhost:3333/"
    private references: any
    private collections: Collection[]
    private websocket: any
    private websocketStatus = {
        'open': WebSocketStatus.OPEN,
        'error': WebSocketStatus.ERROR,
        'close': WebSocketStatus.CLOSE,
    }
    private websocketTimer

    private onChanges: (collection: Collection) => void
    private onErrors: (title: string, msg: string) => void

    public status = WebSocketStatus.NONE

    public collectionSelected: Collection

    constructor (private http: Http) {}

    setOnChanges(changesCb: (collection: Collection) => void) {
        this.onChanges = changesCb
    }

    setOnErrors(errorsCb: (title: string, msg: string) => void) {
        this.onErrors = errorsCb
    }


    selectCollection(collection: Collection) {
        this.collectionSelected = collection
    }

    initWebSocket(): Observable<WebSocketStatus>{
        return Observable.create(
            observer => this.connectWebSocket(observer))
    }

    connectWebSocket(observer) {

        console.log("websocket connecting...")

        // Etablissement de la connexion avec la websocket
        this.websocket = new WebSocket('ws://localhost:3333/ws')

        observer.next(WebSocketStatus.CONNECTING)

        let handleWebSocketStatus = (expected) => {
            return (evt) => {
                let status = this.getWebSocketStatus(evt, expected)

                // V�rification de l'�tat du status
                if (status == undefined) {
                    return
                }

                // V�rification que le status a bien chang�
                if (this.status == expected) {
                    return
                }

                // Attribution du nouveau status
                this.status = expected

                // En cas de status d'erreur ou de fermeture
                // inattendue, lorsque le timer n'est pas d�fini : on
                // relance p�riodiquement la tentative de connexion
                if (this.websocketTimer === undefined
                    && (this.status === WebSocketStatus.ERROR
                        || this.status === WebSocketStatus.CLOSE))  {
                    console.log(
                        "websocket ",
                        this.status === WebSocketStatus.CLOSE ? "close" : "error")
                    this.websocketTimer = setTimeout(
                        () => {
                            console.log("websocket retry ...")
                            this.websocketTimer = undefined
                            this.connectWebSocket(observer)
                        }, 5000)
                }

                observer.next(expected)
            }
        }

        this.websocket.onopen = handleWebSocketStatus(WebSocketStatus.OPEN)
        this.websocket.onerror = handleWebSocketStatus(WebSocketStatus.ERROR)
        this.websocket.onclose = handleWebSocketStatus(WebSocketStatus.CLOSE)
        this.websocket.onmessage = (evt) => {
            console.log("RECEIVED: " + evt.data)
        }
    }

    getWebSocketStatus(evt, expected: WebSocketStatus): WebSocketStatus{

        let status = this.websocketStatus[evt.type]
        if (status == undefined) {
            console.error("Unknown received websocket status type: " + evt.type)
            return undefined
        }

        if (status != expected) {
            console.error("Websocket status error: expected " + expected
                          + ", received " + status)
            return undefined
        }

        return status
    }

    getWebSocket(): Observable<any>{
        return Observable.fromEvent(this.websocket,'message')
    }

    getOptions() {
        return new RequestOptions({
            headers: new Headers({ 'Content-Type': 'application/json' })
        })
    }

    // Create a new collection
    newCollection(collection: Collection) {

        return this.http.post(this.url + "collections",
                              JSON.stringify(collection),
                              this.getOptions())
            .map((res: Response) => {
                if (res.status != 204) {
                    throw new Error('Impossible to create new collection: ' + res.status);
                }

                // Ajoute la collection nouvellement cr��e
                this.collections.push(collection)

                this.onChanges(collection)
            })
            .catch(this.handleError);
    }

    // Modify an existing collection
    modifyCollection(name: string, collection: Collection) {

        return this.http.patch(this.url + "collections/" + name,
                               JSON.stringify(collection),
                               this.getOptions())
            .map((res: Response) => {
                if (res.status != 204) {
                    throw new Error('Impossible to modify collection '
                                    + name + ': ' + res.status);
                }

                // Replace the collection from the list
                for (let i in this.collections) {
                    if (this.collections[i].name === name) {
                        this.collections[i] = collection
                        break
                    }
                }

                // Remove the selected collection
                this.collectionSelected = collection

                this.onChanges(collection)
            })
            .catch(this.handleError);
    }

    // Delete an existing collection
    deleteCollection(name: string) {

        return this.http.delete(this.url + "collections/" + name,
                                this.getOptions())
            .map((res: Response) => {
                if (res.status != 204) {
                    throw new Error('Impossible to modify collection '
                                    + name + ': ' + res.status);
                }

                // Remove the collection from the list
                for (let i = 0; i < this.collections.length; i++) {
                    if (this.collections[i].name === name) {
                        this.collections.splice(i, 1)
                        break
                    }
                }

                // Reset the selected collection
                this.collectionSelected = undefined

                this.onChanges(undefined)
            })
            .catch(this.handleError);
    }

    // Get the collections list
	getAll() {

        return new Observable<Collection[]>(observer => {
            if (this.collections) {
                observer.next(this.collections)
                return
            }

            let request =  this.http.get(this.url + "collections",
                                         this.getOptions())
                .map(this.extractData)
                .catch(this.handleError);

            request.subscribe(collections => {

                if (collections) {
                    this.collections = collections
                    observer.next(collections)
                }
            })
        });
    }


    // Get the collections references
	getReferences() {

        // Setup cache on the references
        return new Observable(observer => {
            if (this.references) {
                observer.next(this.references)
                return
            }

            let request = this.http.get(this.url + "references")
                .map(this.extractData)
                .catch(this.handleError);

            request.subscribe(references => {
                this.references = references
                observer.next(references)
            })
        })
    }


    private extractData(res: Response) {

        if (res.status < 200 || res.status >= 300) {
            throw new Error('Bad response status: ' + res.status);
        }

        // No content to return
        if (res.status === 204) {
            return true
        }

        return res.json();
    }

    private handleError (error: any) {
        let errMsg = error.message || 'Server error';
        if (this.onErrors) {
            this.onErrors("request error", errMsg)
        }

        return Observable.throw(errMsg);
    }
}
