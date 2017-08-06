import { Component, NgZone, OnInit, ViewChild } from '@angular/core';

import { ApiService, CollectionStatus, Event } from './api.service';
import { ImportsService } from './imports/imports.service';
import { BufferService } from './buffer/buffer.service';
import { HomeService } from './home/home.service';
import { ConfigService, ConfigBase } from './config/config.service';

import { Collection } from './collections/collection';

import { CollectionsComponent } from './collections/collections.component'
import { BufferComponent } from './buffer/buffer.component'
import { HomeComponent } from './home/home.component'
import { ItemComponent } from './item/item.component'
import { Item } from './item/item'

declare var jQuery: any;

enum AppStatus {
    NONE = 0,
    HOME,
    IMPORT,
    EXPORT,
    CONFIG,
    ITEM,
}

@Component({
    selector: 'app',
    templateUrl: './app.component.html',
})

export class AppComponent implements OnInit {
    @ViewChild(CollectionsComponent) collections: CollectionsComponent
    @ViewChild(BufferComponent) buffer: BufferComponent
    @ViewChild(HomeComponent) home: HomeComponent

    public appStatus = AppStatus
    public status = AppStatus.NONE

    public title = "Classify"

    public collection: Collection

    public modalTitle: string
    public modalMsg: string

    private importsLoop: any
    private importsRunningNb: number

    @ViewChild(ItemComponent) itemComponent: ItemComponent
    public item: Item

    constructor(private zone: NgZone,
        private apiService: ApiService,
        private importsService: ImportsService,
        private configService: ConfigService,
        private bufferService: BufferService,
        private homeService: HomeService) { }

    ngOnInit() {

        // Initialisation de la side bar
        jQuery(".button-collapse").sideNav();

        // Logo d'importation
        this.importsLoop = jQuery("i#imports-loop")
        this.importsRunningNb = 0;

        // Inscription au flux
        this.apiService.getStream()
            .subscribe((e: Event) => {

                console.log("EVENT!", e)

                // Detect restart application
                if (e.event === "start") {
                    window.location.replace("/");
                    return;
                }

                // Import data
                if (new RegExp("^imports").test(e.event)) {
                    this.handleImport(e);
                    return;
                }

                // Collections reception
                if (new RegExp("^collection").test(e.event)) {
                    this.handleCollection(e);
                    return;
                }
            })

        this.apiService.subscribeCollectionChange(
            (collection: Collection, status: CollectionStatus) => {

                if (collection === undefined) {
                    this.onChangeCollection()
                    return
                }

                this.title = collection.name
                this.collection = collection

                // Get all configuration specific to the collection
                this.configService.getConfigs(collection.name)
                    .subscribe((config: ConfigBase) => { })

                switch (status) {
                    case CollectionStatus.CREATED:
                    case CollectionStatus.MODIFIED:
                    case CollectionStatus.SELECTED:
                        this.onHome()
                        break;
                    case CollectionStatus.DELETED:
                        this.status = AppStatus.NONE
                        break
                }
            })
    }

    onHome() {
        this.onNewState(AppStatus.HOME)
    }

    onImport() {
        this.onNewState(AppStatus.IMPORT)
    }

    onExport() {
        this.onNewState(AppStatus.EXPORT)
    }

    onConfig() {
        this.onNewState(AppStatus.CONFIG)
    }

    onItem(item: Item) {
        this.zone.run(() => {
            this.item = item
            this.onNewState(AppStatus.ITEM)
        })
    }

    onBuffer() {
        this.buffer.start()
    }

    onNewState(nextStatus: AppStatus) {
        this.resetCollectionState()
        this.status = nextStatus
    }

    // Gestion des nouveaux imports
    handleImport(e: Event) {
        // Send notifications to the imports list
        this.importsService.addEvent(e);

        // Display imports status
        if (new RegExp('status$').test(e.event)) {

            // Status 'TRUE': rotate refresh logo
            if (e.data) {
                this.importsLoop.addClass("rotation")
                this.importsRunningNb++
            }
            // Status 'FALSE'
            else if (this.importsRunningNb > 0) {
                this.importsRunningNb--
            }

            // No more imports : stop logo rotation
            if (this.importsRunningNb < 1) {
                this.importsLoop.removeClass("rotation")
            }
        }
    }

    handleCollection(e: Event) {

        let names = e.event.split("/");
        let size = names.length;

        if (size <= 1) {
            console.error("Invalid collection event '" + e.event + "'")
            return;
        }

        let collection = names[1];
        let destination: string
        if (size > 2) {
            destination = names[2];
        }

        let item = new Item(e.data)

        if (this.item != undefined && this.item.id == item.id) {
            this.itemComponent.onUpdate(item)
            this.item = item
        }

        // Send notifications to the imports list
        if (destination === "buffer") {
            this.bufferService.addEvent(collection, e, item)
        }
        else if (destination === "items") {
            this.homeService.addEvent(collection, e, item)
        }

    }

    onError(title: string, msg: string) {

        console.error(title + "error :" + msg)

        this.modalTitle = title + " error!"
        this.modalMsg = msg
        jQuery('#modal').openModal()
    }

    stopModal() {
        jQuery('#modal').closeModal()
    }

    // Affiche la création d'une nouvelle collection
    onNewCollection() {
        if (this.collections) {
            this.collections.onNewCollection()
        }
    }

    // Affiche la liste des collections à sélectionner
    onChangeCollection() {
        if (this.collections
            && this.collections.onChooseCollection(undefined)) {
            this.status = AppStatus.NONE
        }
    }

    // Affiche la liste des collections à sélectionner si aucune
    // collection n'est actuellement sélectionnée
    resetCollectionState() {
        if (this.collections)
            this.collections.resetCollectionState()
    }
}
